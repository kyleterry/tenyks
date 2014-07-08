package irc

import (
	"container/list"
	"fmt"
)

type handlerRegistry struct {
	handlers map[string]*list.List
}

type fn func (*Connection, *Message)

func (self *Connection) AddHandler(name string, handler fn) {
	if _, ok := self.Registry.handlers[name]; !ok {
		self.Registry.handlers[name] = list.New()
	}
	self.Registry.handlers[name].PushBack(handler)
}

func (self *Connection) addBaseHandlers () {
	self.AddHandler("bootstrap", (*Connection).BootstrapHandler)
	self.AddHandler("001", (*Connection).ConnectedHandler)
	self.AddHandler("433", (*Connection).NickInUseHandler)
	self.AddHandler("PING", (*Connection).PingHandler)
}

func (self *Connection) PingHandler(msg *Message) {
	self.Out <- fmt.Sprintf("PONG %s", msg.Trail)
}

func (self *Connection) BootstrapHandler(msg *Message) {
	self.Out <- fmt.Sprintf(
		"USER %s %s %s :%s",
		self.Config.Nicks[0],
		self.Config.Host,
		self.Config.Ident,
		self.Config.Realname)
	self.Out <- fmt.Sprintf(
		"NICK %s", self.Config.Nicks[self.nickIndex])
	self.currentNick = self.Config.Nicks[self.nickIndex]
}

func (self *Connection) NickInUseHandler(msg *Message) {
	self.nickIndex++
	if len(self.Config.Nicks) >= self.nickIndex + 1 {
		self.Out <- fmt.Sprintf(
			"NICK %s", self.Config.Nicks[self.nickIndex])
		self.currentNick = self.Config.Nicks[self.nickIndex]
	} else {
		log.Fatal("All nicks in use.")
	}
}

func (self *Connection) ConnectedHandler(msg *Message) {
	for _, channel := range self.Config.Channels {
		self.Out <- fmt.Sprintf("JOIN %s", channel)
	}
}
