package service

import (
	"encoding/json"
	"time"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
)

type ServiceEngine struct {
	Reactor   *PubSubReactor
	ServiceRg *ServiceRegistry
	CommandRg *irc.HandlerRegistry
	ircconns  irc.IrcConnections
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
	self.addBaseHandlers()
	self.Reactor.Start()
	go self.NotifyServicesAboutStart()
	go self.serviceWatchdog()
	go self.Reactor.conn.PingServices()
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

func (self *ServiceEngine) serviceWatchdog() {
	for {
		<-time.After(time.Second * 180)
		log.Debug("[service] Checking service health")
		services := self.ServiceRg.services
		for name, service := range services {
			if service.Online {
				log.Debug("[service] Checking %s", name)
				pongDuration := time.Since(service.LastPong)
				testDuration := time.Duration(time.Second * 400)
				if int(pongDuration) > int(testDuration) {
					log.Debug("[service] %s seems to have gone offline", name)
					service.Online = ServiceOffline
				}
			}
		}
	}
}

// UpdateService is used to update the service state.
func (self *ServiceEngine) UpdateService(uuid string, status bool) {
	self.ServiceRg.regMu.Lock()
	defer self.ServiceRg.regMu.Unlock()
	if _, ok := self.ServiceRg.services[uuid]; ok {
		service := self.ServiceRg.services[uuid]
		service.Online = status
		if status {
			log.Debug("[service] %s responded with PONG",
				service.UUID.String())
			service.LastPong = time.Now()
		}
	}
}

// NotifyServicesAboutStart that the bot came online. This should prompt
// services to send the REGISTER command.
func (self *ServiceEngine) NotifyServicesAboutStart() {
	log.Debug("[service] Notifying Services that Tenyks came online")
	msg := &Message{
		Command: "HELLO",
		Payload: "!tenyks",
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		log.Fatal(err)
	}
	self.Reactor.conn.Out <- string(jsonBytes[:])
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
