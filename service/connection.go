package service

import (
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
	"github.com/kyleterry/tenyks/config"
)

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

func (self *Connection) Bootstrap() {
	self.In = self.recv()
	self.Out = self.send()
	self.pubsub = redis.PubSubConn{self.r}
}

func (self *Connection) recv() <-chan []byte {
	c := make(chan []byte, 1000)
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
