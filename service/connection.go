package service

import (
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/op/go-logging"
	zmq "github.com/pebbe/zmq4"
)

var log = logging.MustGetLogger("tenyks")

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
		return nil, err
	}
	sender, err := ctx.NewSocket(zmq.PUB)
	if err != nil {
		return nil, err
	}
	receiver, err := ctx.NewSocket(zmq.SUB)
	if err != nil {
		return nil, err
	}
	receiver.SetSubscribe("")

	conn.pubsub.ctx = ctx
	conn.pubsub.sender = sender
	conn.pubsub.receiver = receiver
	return conn, nil
}

func (c *Connection) Init() {
	// Hook up PrivmsgHandler to all connections
	log.Debug("[service] Bootstrapping pubsub")
	c.pubsub.sender.Bind(c.config.SenderBind)
	log.Debug("[service] sender is listening on %s", c.config.SenderBind)
	c.pubsub.receiver.Bind(c.config.ReceiverBind)
	log.Debug("[service] receiver is listening on %s", c.config.ReceiverBind)
	c.In = c.recv()
	c.Out = c.send()
}

// RegisterIRCHandlers will is what connections handler functions to IRC
// connection instances.
func (c *Connection) RegisterIrcHandlers(conn *irc.Connection) {
	log.Debug("[service] Registring IRC Handlers")
	conn.AddHandler("PRIVMSG", c.PrivmsgIrcHandler)
	conn.AddHandler("PRIVMSG", c.ListServicesIrcHandler)
	conn.AddHandler("PRIVMSG", c.HelpIrcHandler)
	conn.AddHandler("PRIVMSG", c.InfoIrcHandler)
}

func (c *Connection) recv() <-chan string {
	ch := make(chan string, 1000)
	log.Debug("[service] Starting recv loop")
	go func() {
		for {
			msgs, err := c.pubsub.receiver.RecvMessage(0)
			log.Debug("[service] recv loop: received message")
			if err != nil {
				log.Debug("[service] recv loop: message error (%s)", err)
				continue
			}
			for _, msg := range msgs {
				ch <- msg
			}
		}
	}()
	return ch
}

func (c *Connection) publish(msg string) {
	c.pubsub.sender.SendMessage(msg)
}

func (c *Connection) getIrcConnByName(name string) *irc.Connection {
	conn, ok := c.engine.ircconns[name]
	if !ok {
		log.Error("[service] Connection `%s` doesn't exist", name)
	}
	return conn
}

func (c *Connection) send() chan<- string {
	ch := make(chan string, 1000)
	log.Debug("[service] Spawning send loop")
	go func() {
		for {
			select {
			case msg := <-ch:
				c.publish(msg)
			}
		}
	}()
	return ch
}
