package service

import (
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
)

type ServiceEngine struct {
	Reactor  *PubSubReactor
	ServiceRg *ServiceRegistry
	CommandRg *irc.HandlerRegistry
	ircconns irc.IrcConnections
}

func NewServiceEngine(conf config.RedisConfig) *ServiceEngine {
	eng := &ServiceEngine{}
	eng.Reactor = NewPubSubReactor(conf)
	eng.Reactor.engine = eng
	eng.Reactor.conn.engine = eng
	eng.ServiceRg = NewServiceRegistry()
	eng.CommandRg = irc.NewHandlerRegistry()
	return eng
}

func (self *ServiceEngine) Start() {
	log.Info("[service] Starting engine")
	self.Reactor.Start()
}

func (self *ServiceEngine) SetIrcConns(ircconns irc.IrcConnections) {
	self.ircconns = ircconns
}

func (self *ServiceEngine) RegisterIrcHandlersFor(conn *irc.Connection) {
	// wait for connect wait to return true and or be closed before registring
	// PRIVMSG handlers. This is to prevent race conditions.
	if val, ok := <-conn.ConnectWait; val == true || !ok { // Wait for connect
		self.Reactor.conn.RegisterIrcHandlers(conn)
	}
}

type PubSubReactor struct {
	conn   *Connection
	engine *ServiceEngine
}

func NewPubSubReactor(conf config.RedisConfig) *PubSubReactor {
	reactor := &PubSubReactor{}
	reactor.conn = NewConn(conf)
	reactor.conn.Bootstrap()
	return reactor
}

func (self *PubSubReactor) Start() {
	log.Debug("[service] Starting Pub/Sub reactor")
	go func() {
		for {
			select {
			case msg := <-self.conn.In:
				go self.conn.dispatch(msg)
			}
		}
	}()
}
