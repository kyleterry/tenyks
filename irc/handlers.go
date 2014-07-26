package irc

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type HandlerRegistry struct {
	Handlers   map[string]*list.List
	RegistryMu *sync.Mutex
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		Handlers: make(map[string]*list.List),
		RegistryMu: &sync.Mutex{},
	}
}

func (self *HandlerRegistry) AddHandler(name string, handler *Handler) {
	self.RegistryMu.Lock()
	defer self.RegistryMu.Unlock()
	if _, ok := self.Handlers[name]; !ok {
		self.Handlers[name] = list.New()
	}
	self.Handlers[name].PushBack(handler)
}

type Handler struct {
	Fn func(...interface{})
}

func NewHandler(fn func(...interface{})) *Handler {
	return &Handler{fn}
}

type ircfn func(*Connection, *Message)

func (self *Connection) AddHandler(name string, fn ircfn) {
	handler := NewHandler(func(p ...interface{}) {
		fn(p[0].(*Connection), p[1].(*Message))
	})
	self.Registry.AddHandler(name, handler)
}

func (self *Connection) addBaseHandlers() {
	self.AddHandler("bootstrap", (*Connection).BootstrapHandler)
	self.AddHandler("send_ping", (*Connection).SendPing)
	self.AddHandler("001", (*Connection).ConnectedHandler)
	self.AddHandler("433", (*Connection).NickInUseHandler)
	self.AddHandler("PING", (*Connection).PingHandler)
	self.AddHandler("PONG", (*Connection).PongHandler)
	self.AddHandler("CTCP", (*Connection).CTCPHandler)
}

func (self *Connection) PingHandler(msg *Message) {
	log.Debug("[%s] Responding to PING", self.Name)
	self.Out <- fmt.Sprintf("PONG %s", msg.Trail)
}

func (self *Connection) PongHandler(msg *Message) {
	self.LastPong = time.Now()
	self.PongIn <- true
}

func (self *Connection) SendPing(msg *Message) {
	log.Debug("[%s] Sending PING to server %s", self.Name, self.currentServer)
	self.Out <- fmt.Sprintf("PING %s", self.currentServer)
}

func (self *Connection) BootstrapHandler(msg *Message) {
	log.Info("[%s] Bootstrapping connection", self.Name)
	self.Out <- fmt.Sprintf(
		"USER %s %s %s :%s",
		self.Config.Nicks[0],
		self.Config.Host,
		self.Config.Ident,
		self.Config.Realname)
	self.Out <- fmt.Sprintf(
		"NICK %s", self.Config.Nicks[self.nickIndex])
	self.currentNick = self.Config.Nicks[self.nickIndex]
	self.ConnectWait <- true
	close(self.ConnectWait)
}

func (self *Connection) NickInUseHandler(msg *Message) {
	log.Info("[%s] Nick `%s` is in use. Next...", self.Name, self.currentNick)
	self.nickIndex++
	if len(self.Config.Nicks) >= self.nickIndex+1 {
		self.Out <- fmt.Sprintf(
			"NICK %s", self.Config.Nicks[self.nickIndex])
		self.currentNick = self.Config.Nicks[self.nickIndex]
	} else {
		log.Fatal("All nicks in use.")
	}
}

func (self *Connection) ConnectedHandler(msg *Message) {
	log.Info("[%s] Sending user commands", self.Name)
	for _, commandHook := range self.Config.Commands {
		self.Out <- commandHook
	}
	log.Info("[%s] Joining Channels", self.Name)
	for _, channel := range self.Config.Channels {
		self.Out <- fmt.Sprintf("JOIN %s", channel)
		log.Debug("[%s] Joined %s", self.Name, channel)
	}
	self.currentServer = msg.Prefix
	go self.watchdog()
}

func (self *Connection) CTCPHandler(msg *Message) {

}
