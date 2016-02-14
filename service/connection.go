package service

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Connection struct {
	r      redis.Conn
	config *config.RedisConfig
	In     <-chan []byte
	Out    chan<- string
	pubsub redis.PubSubConn
	engine *ServiceEngine
}

func NewConn(conf config.RedisConfig) *Connection {
	redisAddr := fmt.Sprintf(
		"%s:%d",
		conf.Host,
		conf.Port)
	r, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		log.Fatal(err)
	}
	conn := &Connection{
		r:      r,
		config: &conf,
	}
	return conn
}

func (self *Connection) Bootstrap() {
	// Hook up PrivmsgHandler to all connections
	log.Debug("[service] Bootstrapping pubsub")
	self.pubsub = redis.PubSubConn{Conn: self.r}
	self.In = self.recv()
	self.Out = self.send()
}

// RegisterIRCHandlers will is what connections handler functions to IRC
// connection instances.
func (self *Connection) RegisterIrcHandlers(conn *irc.Connection) {
	log.Debug("[service] Registring IRC Handlers")
	conn.AddHandler("PRIVMSG", self.PrivmsgIrcHandler)
	conn.AddHandler("PRIVMSG", self.ListServicesIrcHandler)
	conn.AddHandler("PRIVMSG", self.HelpIrcHandler)
	conn.AddHandler("PRIVMSG", self.InfoIrcHandler)
}

func (self *Connection) DialRedis() (redis.Conn, error) {
	redisAddr := fmt.Sprintf(
		"%s:%d",
		self.config.Host,
		self.config.Port)
	r, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return nil, err
	}
	return r, nil
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
	c, err := self.DialRedis()
	if err != nil {
		panic(err)
	}
	defer c.Close()
	c.Do("PUBLISH", channel, msg)
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
