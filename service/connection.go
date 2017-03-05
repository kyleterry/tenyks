package service

import (
	"fmt"
	"strings"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	zmq "github.com/pebbe/zmq4"
	"github.com/pkg/errors"
)

type pubsub struct {
	ctx      *zmq.Context
	sender   *zmq.Socket
	receiver *zmq.Socket
}

type Connection struct {
	config *config.ServiceConfig
	In     <-chan string
	Out    chan<- string
	pubsub *pubsub
	engine *ServiceEngine
}

func NewConn(conf config.ServiceConfig) (*Connection, error) {
	conn := &Connection{
		config: &conf,
		pubsub: &pubsub{},
	}
	ctx, err := zmq.NewContext()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create zeromq context")
	}
	sender, err := ctx.NewSocket(zmq.PUB)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create zeromq pub socket")
	}
	receiver, err := ctx.NewSocket(zmq.SUB)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create zeromq sub socket")
	}
	receiver.SetSubscribe("")

	conn.pubsub.ctx = ctx
	conn.pubsub.sender = sender
	conn.pubsub.receiver = receiver
	return conn, nil
}

func (c *Connection) Init() error {
	// Hook up PrivmsgHandler to all connections
	var sbind string
	if strings.Contains(c.config.SenderBind, "tcp://") {
		sbind = c.config.SenderBind
	} else {
		sbind = fmt.Sprintf("tcp://%s", c.config.SenderBind)
	}
	err := c.pubsub.sender.Bind(sbind)
	if err != nil {
		return errors.Wrap(err, "cannot bind sender")
	}
	Logger.Info("sender is now listening", "addr", sbind)
	var rbind string
	if strings.Contains(c.config.ReceiverBind, "tcp://") {
		rbind = c.config.ReceiverBind
	} else {
		rbind = fmt.Sprintf("tcp://%s", c.config.ReceiverBind)
	}
	err = c.pubsub.receiver.Bind(rbind)
	if err != nil {
		return errors.Wrap(err, "cannot bind receiver")
	}
	Logger.Info("receiver is now listening", "addr", rbind)
	c.In = c.recv()
	c.Out = c.send()

	return nil
}

// RegisterIrcHandlers connects a few callback handlers for PRIVMSG's that respond
// to things like !services and !help
func (c *Connection) RegisterIrcHandlers(conn *irc.Connection) {
	Logger.Debug("registring IRC Handlers")
	conn.AddHandler("PRIVMSG", c.PrivmsgIrcHandler)
	conn.AddHandler("PRIVMSG", c.ListServicesIrcHandler)
	conn.AddHandler("PRIVMSG", c.HelpIrcHandler)
	conn.AddHandler("PRIVMSG", c.InfoIrcHandler)
}

func (c *Connection) recv() <-chan string {
	ch := make(chan string, 1000)
	Logger.Debug("starting recv loop")
	go func() {
		for {
			msgs, err := c.pubsub.receiver.RecvMessage(0)
			if err != nil {
				Logger.Debug("received malformed message", "error", err, "msg", msgs)
				continue
			}
			for _, msg := range msgs {
				ch <- msg
			}
		}
	}()
	return ch
}

func (c *Connection) publish(msg string) error {
	_, err := c.pubsub.sender.SendMessage(msg)
	if err != nil {
		return errors.Wrap(err, "could not send message")
	}

	return nil
}

func (c *Connection) getIrcConnByName(name string) *irc.Connection {
	if conn, ok := c.engine.ircconns[name]; ok {
		return conn
	}
	return nil
}

func (c *Connection) send() chan<- string {
	ch := make(chan string, 1000)
	Logger.Debug("starting send loop")
	go func() {
		for {
			select {
			case msg := <-ch:
				err := c.publish(msg)
				if err != nil {
					Logger.Error("failed to send message", "error", err)
				}
			}
		}
	}()
	return ch
}
