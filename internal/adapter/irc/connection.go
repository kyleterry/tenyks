package irc

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

	"github.com/kyleterry/tenyks/internal/adapter"
	"github.com/kyleterry/tenyks/internal/logger"
	servicepb "github.com/kyleterry/tenyks/internal/service"
)

var (
	DefaultLoginFunc = defaultLoginFunc
	DefaultJoinFunc  = defaultJoinFunc
)

type OnConnectHook func(context.Context, *Connection) error
type OnDisconnectHook func(context.Context, *Connection) error
type OnCommandHook func(context.Context, *Connection, Command) error
type OnReplyHook func(context.Context, *Connection, Reply) error
type OnErrorHook func(context.Context, *Connection, error) error
type ConnectionCommandFactoryFunc func(*Connection, *Message) Command
type ConnectionReplyFactoryFunc func(*Connection, *Message) Command

type Config struct {
	Name     string
	Server   string
	Password string
	User     string
	RealName string
	UseTLS   bool
	Logger   logger.Logger
	Nicks    []string
	Channels []string
	Commands []string
}

type ConnectionStatus struct {
	Connected               bool
	CurrentNick             string
	CurrentServer           string
	StartedAt               time.Time
	LastServerProbe         time.Time
	LastServerProbeResponse time.Time
}

type Connection struct {
	Name   string
	Status ConnectionStatus

	// event hooks
	OnConnect    []OnConnectHook
	OnDisconnect []OnDisconnectHook
	OnCommand    []OnCommandHook
	OnReply      []OnReplyHook
	OnError      []OnErrorHook

	// factories
	CommandFactory map[CommandType]ConnectionCommandFactoryFunc
	ReplyFactory   map[ReplyType]ConnectionReplyFactoryFunc

	channels map[string]*Channel

	// configuration
	server   string
	password string
	user     string
	realName string
	useTLS   bool
	nicks    []string
	log      logger.Logger

	// managed state
	conn                net.Conn
	cancel              context.CancelFunc
	io                  *bufio.ReadWriter
	in                  chan MessageObject
	out                 chan Command
	chatMessageHandlers []adapter.HandlerFunc

	sync.RWMutex
}

func (c *Connection) GetName() string {
	return c.Name
}

func (c *Connection) GetType() adapter.AdapterType {
	return adapter.AdapterTypeIRC
}

func (c *Connection) RegisterMessageHandler(h adapter.HandlerFunc) {
	c.chatMessageHandlers = append(c.chatMessageHandlers, h)
}

func (c *Connection) Dial(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	b := backoff{max: time.Minute * 5}

	for {
		if c.conn != nil {
			break
		}

		dialer := net.Dialer{}

		conn, err := dialer.DialContext(ctx, "tcp", c.server)
		if err != nil {
			dur := b.next()
			c.log.Error("connection failed",
				logger.Param{Key: "connection", Value: c.Name},
				logger.Param{Key: "retry", Value: dur})

			<-time.After(dur)

			continue
		}

		c.conn = conn

		c.io = bufio.NewReadWriter(
			bufio.NewReader(c.conn),
			bufio.NewWriter(c.conn),
		)

		in, recvErrs := c.startReceiveLoop(ctx)
		out, sendErrs := c.startSendLoop(ctx)

		c.in = in
		c.out = out

		// TODO replace with OnError hooks
		go c.monitorErrs(ctx, recvErrs, sendErrs)
		go c.temporaryDispatcher(ctx)

		for _, hook := range c.OnConnect {
			if err := hook(ctx, c); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Connection) Close(ctx context.Context) error {
	if c.Status.Connected {
		c.conn.Close()
		c.conn = nil

		c.cancel()

		close(c.in)
		close(c.out)

		c.Status.Connected = false

		for _, hook := range c.OnDisconnect {
			if err := hook(ctx, c); err != nil {
				return err
			}
		}
	}

	return nil
}

// EnqueueCommand takes a Command, calls its Validate method and puts it on the
// send queue (out channel). If the out channel's buffer is full, this method
// will block until some commands are flushed by the io workers and buffer
// slots are freed up.
func (c *Connection) EnqueueCommand(cmd Command) error {
	if err := cmd.Validate(); err != nil {
		return fmt.Errorf("failed to enqueue command; validation failed: %w", err)
	}

	c.out <- cmd

	return nil
}

// SendAsync takes a generic tenyks message and decodes it into a PRIVCMD and
// attempts to put it on the send queue. See EnqueueCommand for information on
// potential contention.
func (c *Connection) SendAsync(_ context.Context, msg *servicepb.Message) error {
	decoder := tenyksChatMessageDecoder{}
	cmd, err := decoder.Decode(msg)
	if err != nil {
		return fmt.Errorf("failed to send message; decoding failed: %w", err)
	}

	return c.EnqueueCommand(cmd)
}

// WithWriteLock returns this connection with it's mutex locked for writing.
// The lock will be released when the callback fn returns.
func (c *Connection) WithWriteLock(_ context.Context, fn func(*Connection)) error {
	c.Lock()
	defer c.Unlock()

	fn(c)

	return nil
}

func (c *Connection) WithReadLock(_ context.Context, fn func(*Connection)) error {
	c.RLock()
	defer c.RUnlock()

	fn(c)

	return nil
}

func (c *Connection) monitorErrs(ctx context.Context, in chan error, out chan error) {
	var err error

	for {
		select {
		case err = <-in:
			c.log.Error("receive error", logger.Param{Key: "error", Value: err})
		case err = <-out:
			c.log.Error("send error", logger.Param{Key: "error", Value: err})
		case <-ctx.Done():
			return
		}

		for _, hook := range c.OnError {
			hook(ctx, c, err)
		}
	}
}

func (c *Connection) startSendLoop(ctx context.Context) (chan Command, chan error) {
	outCh := make(chan Command, 10)
	errCh := make(chan error)

	go func() {
		for {
			select {
			case cmd, ok := <-outCh:
				if !ok {
					return
				}

				if err := cmd.Validate(); err != nil {
					errCh <- err

					continue
				}

				msg, err := cmd.Encode()
				if err != nil {
					errCh <- err

					continue
				}

				c.log.Debug(cmd.Message().RawMsg, logger.Param{Key: "direction", Value: "|--->|"})

				if _, err := c.io.WriteString(msg); err != nil {
					errCh <- err

					continue
				}

				if err := c.io.Flush(); err != nil {
					errCh <- err

					continue
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	return outCh, errCh
}

func (c *Connection) startReceiveLoop(ctx context.Context) (chan MessageObject, chan error) {
	inCh := make(chan MessageObject, 10)
	errCh := make(chan error)

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				line, err := c.io.ReadString('\n')
				if err != nil {
					errCh <- err

					continue
				}

				mo, err := c.decodeAndMapMessage(line)
				if err != nil {
					errCh <- err
					continue
				}

				c.log.Debug(mo.Message().RawMsg, logger.Param{Key: "direction", Value: "|<---|"})

				inCh <- mo
			}
		}
	}()

	return inCh, errCh
}

func (c *Connection) decodeAndMapMessage(raw string) (MessageObject, error) {
	msg, err := NewRawMessageDecoder().Decode(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	var mo MessageObject
	mo = &UnknownCommand{m: msg}

	if msg.MessageType == MessageTypeCommand {
		if ct, ok := CommandTypeMapping[msg.Command]; ok {
			// first we check the connection's command factory, if there's no
			// mapping, then we move on to the default one.
			if fn, ok := c.CommandFactory[ct]; ok {
				mo = fn(c, msg)
			} else if fn, ok := DefaultCommandFactory[ct]; ok {
				mo = fn(msg)
			}
		}
	} else {
		if rt, ok := ReplyTypeMapping[msg.Command]; ok {
			// first we check the connection's reply factory, if there's no
			// mapping, then we move on to the default one.
			if fn, ok := c.ReplyFactory[rt]; ok {
				mo = fn(c, msg)
			} else if fn, ok := DefaultReplyFactory[rt]; ok {
				mo = fn(msg)
			}
		}
	}

	return mo, nil
}

func (c *Connection) temporaryDispatcher(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case mo := <-c.in:
				switch mo.Message().MessageType {
				case MessageTypeCommand:
					for _, hook := range c.OnCommand {
						hook(ctx, c, mo.(Command))
					}
				case MessageTypeReply:
					for _, hook := range c.OnReply {
						hook(ctx, c, mo.(Reply))
					}
				}
			}
		}
	}()
}

func New(conf Config) (*Connection, error) {
	if _, _, err := net.SplitHostPort(conf.Server); err != nil {
		return nil, err
	}

	channels := map[string]*Channel{}

	for _, channel := range conf.Channels {
		channels[channel] = NewChannel(channel)
	}

	return &Connection{
		Name: conf.Name,
		OnConnect: []OnConnectHook{
			defaultLoginFunc,
			defaultJoinFunc,
		},
		OnDisconnect: []OnDisconnectHook{
			cleanupChannels,
		},
		OnCommand: []OnCommandHook{
			defaultJoinChannelStatusUpdater,
			defaultPrivmsgHandler,
			defaultUnknownHandler,
			defaultPingResponder,
		},
		OnReply: []OnReplyHook{
			defaultConnectionStatusUpdater,
			defaultChannelMemberUpdater,
		},
		OnError: []OnErrorHook{
			defaultErrorHandlerFunc,
		},
		CommandFactory: map[CommandType]ConnectionCommandFactoryFunc{
			CommandTypePrivmsg: mentionAndDirectPrivmsgCommand,
		},
		Status: ConnectionStatus{
			StartedAt: time.Now(),
		},
		server:   conf.Server,
		useTLS:   conf.UseTLS,
		channels: channels,
		nicks:    conf.Nicks,
		user:     conf.User,
		realName: conf.RealName,
		password: conf.Password,
		log:      conf.Logger,
	}, nil
}

type backoff struct {
	factor float64
	min    time.Duration
	max    time.Duration

	n uint64
}

func (b *backoff) next() time.Duration {
	factor := b.factor

	if factor < 1 {
		factor = 2
	}

	d := float64(b.min) * math.Pow(factor, float64(b.n))

	if d > float64(b.max) {
		d = float64(b.max)
	} else {
		b.n++
	}

	return time.Duration(d)
}
