package service

import (
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
)

type ServiceEngine struct {
	Reactor *PubSubReactor
	Registry *ServiceRegistry
}

func NewServiceEngine(conf config.RedisConfig, ircconns irc.IrcConnections) *ServiceEngine {
	eng := &ServiceEngine{}
	eng.Reactor = NewPubSubReactor(conf, ircconns)
	eng.Registry = NewServiceRegistry()
	return eng
}

func (self *ServiceEngine) Start() {
	log.Info("[service] Starting engine")
	go self.Reactor.Start()
}

type PubSubReactor struct {
	conn *Connection
	ircconns irc.IrcConnections
}

func NewPubSubReactor(conf config.RedisConfig, ircconns irc.IrcConnections) *PubSubReactor {
	reactor := &PubSubReactor{}
	reactor.ircconns = ircconns
	reactor.conn = NewConn(conf, reactor.ircconns)
	reactor.conn.Bootstrap()
	return reactor
}

func (self *PubSubReactor) Start() {
	self.conn.registerIrcHandlers()
	log.Debug("[service] Starting Pub/Sub reactor")
	go func(){
		for {
			select {
			case msg := <-self.conn.In:
				go self.conn.dispatch(msg)
			}
		}
	}()
}
