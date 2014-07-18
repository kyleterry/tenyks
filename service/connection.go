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

func (self *Connection) Bootstrap(ircconns *map[string]*irc.Connection) {
	log.Debug("[service] Bootstrapping pubsub")
	self.In = self.recv()
	self.Out = self.send()
	self.pubsub = redis.PubSubConn{self.r}

	// Hook up PrivmsgHandler to all connections
	for _, ircconn := range *ircconns {
		ircconn.AddHandler("PRIVMSG", PrivmsgHandler)
	}
}

func (self *Connection) recv() <-chan []byte {
	c := make(chan []byte, 1000)
	log.Debug("[service] Spawning recv loop")
	go func() {
		self.pubsub.Subscribe(self.config.TenyksPrefix + ".broadcast")
		for {
			switch msg := self.pubsub.Receive().(type) {
			case redis.Message:
				c <- msg.Data
			}
		}
	}()
	return c
}

func (self *Connection) send() chan<- string {
	c := make(chan string, 1000)
	log.Debug("[service] Spawning send loop")
	go func() {
		for {
			select {
			case msg := <-c:
				self.r.Send("PUBLISH", msg)
			}
		}
	}()
	return c
}
