package service

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Connection struct {
	config *config.ServiceConfig
	In     <-chan []byte
	Out    chan<- string
	publisher *zmq.Socket
	engine *ServiceEngine
}

func NewConn(conf *config.ServiceConfig) *Connection {
	conn := &Connection{
		config: &conf,
	}
	return conn
}

func (self *Connection) Bootstrap() {
	// Hook up PrivmsgHandler to all connections
	log.Debug("[service] Bootstrapping pubsub")
	self.publisher = zmq.NewSocket(zmq.PUB)
	self.publisher.Bind("tcp://*:6669")
	self.In = self.recv()
	self.Out = self.send()
}

// RegisterIRCHandlers will is what connections handler functions to IRC
// connection instances.
func (self *Connection) RegisterIrcHandlers(conn *irc.Connection) {
	log.Debug("[service] Registring IRC Handlers")
	log.Debug("[service] Registring PRIVMSG handler with `%s`", conn.Name)
	conn.AddHandler("PRIVMSG", self.PrivmsgIrcHandler)
}

func (self *Connection) recv() <-chan []byte {
	c := make(chan []byte, 1000)
	log.Debug("[service] Spawning recv loop")
	go func() {
		self.pubsub.Subscribe(self.getTenyksChannel())
		for {
			switch msg := self.pubsub.Receive().(type) {
			case redis.Message:
				c <- msg.Data
			}
		}
	}()
	return c
}

func (self *Connection) publish(channel, msg string) {
	_, err := self.publisher.SendMessage(msg)
	if err != nil {
		log.Error(err)
	}
}

func (self *Connection) getServiceChannel() string {
	channel := "tenyks.service.broadcast"
	if self.config.ServiceChannel != "" {
		channel = self.config.ServiceChannel
	}
	return channel
}

func (self *Connection) getTenyksChannel() string {
	channel := "tenyks.robot.broadcast"
	if self.config.TenyksChannel != "" {
		channel = self.config.TenyksChannel
	}
	return channel
}

func (self *Connection) getIrcConnByName(name string) *irc.Connection {
	conn, ok := self.engine.ircconns[name]
	if !ok {
		log.Error("[service] Connection `%s` doesn't exist", name)
	}
	return conn
}

func (self *Connection) send() chan<- string {
	c := make(chan string, 1000)
	log.Debug("[service] Spawning send loop")
	go func() {
		for {
			select {
			case msg := <-c:
				self.publish(self.getServiceChannel(), msg)
			}
		}
	}()
	return c
}
