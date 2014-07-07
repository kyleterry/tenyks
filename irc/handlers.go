package irc

import (
	"container/list"
	"fmt"
)

type handlerRegistry struct {
	handlers map[string]*list.List
}

type fn func (*Connection, *Message)

var ircHandlers handlerRegistry

func addBaseHandlers () {
	ircHandlers.handlers["PING"].PushBack((*Connection).PingHandler)
}

func (self *Connection) PingHandler(msg *Message) {
	self.Out <- fmt.Sprintf("PONG %s", msg.Trail)
}
