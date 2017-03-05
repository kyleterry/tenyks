package service

import (
	"encoding/json"
	"time"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/pkg/errors"
)

type ServiceEngine struct {
	Reactor   *PubSubReactor
	ServiceRg *ServiceRegistry
	CommandRg *irc.HandlerRegistry
	ircconns  irc.IRCConnections
}

func NewServiceEngine(conf config.ServiceConfig) (*ServiceEngine, error) {
	var err error

	eng := &ServiceEngine{}
	eng.Reactor, err = NewPubSubReactor(conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pub/sub reactor")
	}
	eng.Reactor.engine = eng
	eng.Reactor.conn.engine = eng
	eng.ServiceRg = NewServiceRegistry()
	eng.CommandRg = irc.NewHandlerRegistry()
	return eng, nil
}

func (self *ServiceEngine) Start() {
	Logger.Info("starting engine")
	self.addBaseHandlers()
	self.Reactor.Start()
	// TODO: these need to take error channels so we can handle things that go wrong
	go self.NotifyServicesAboutStart()
	go self.serviceWatchdog()
	go self.Reactor.conn.PingServices()
}

func (self *ServiceEngine) SetIRCConns(ircconns irc.IRCConnections) {
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
		services := self.ServiceRg.services
		for name, service := range services {
			if service.Online {
				Logger.Debug("checking service health", "service", name)
				pongDuration := time.Since(service.LastPong)
				testDuration := time.Duration(time.Second * 400)
				if int(pongDuration) > int(testDuration) {
					Logger.Debug("marking service offline", "service", name)
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
			Logger.Debug("service responded to PING", "service-id", service.UUID.String())
			service.LastPong = time.Now()
		}
	}
}

// NotifyServicesAboutStart that the bot came online. This should prompt
// services to send the REGISTER command.
func (self *ServiceEngine) NotifyServicesAboutStart() {
	Logger.Debug("notifying services of startup")
	msg := &Message{
		Command: "HELLO",
		Payload: "!tenyks",
	}
	jsonBytes, err := json.Marshal(msg)
	if err != nil {
		Logger.Error("failed to serialize msg", "error", err)
	}
	self.Reactor.conn.Out <- string(jsonBytes[:])
}

type PubSubReactor struct {
	conn   *Connection
	engine *ServiceEngine
}

func NewPubSubReactor(conf config.ServiceConfig) (*PubSubReactor, error) {
	reactor := &PubSubReactor{}
	conn, err := NewConn(conf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create connection object")
	}
	reactor.conn = conn
	if err := reactor.conn.Init(); err != nil {
		return nil, errors.Wrap(err, "failed to initialize service connection")
	}
	return reactor, nil
}

func (p *PubSubReactor) Start() {
	Logger.Debug("starting pub/sub reactor")
	go func() {
		for {
			select {
			case msg := <-p.conn.In:
				go p.conn.dispatch(msg)
			}
		}
	}()
}
