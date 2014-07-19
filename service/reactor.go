package service

import (
	"github.com/kyleterry/tenyks/irc"
)

func ConnectionReactor(ircconns map[string]*irc.Connection,
	conn *Connection) {
	conn.Bootstrap(ircconns)
	for {
		select {
		case msg := <-conn.In:
			go conn.dispatch(msg)
		}
	}
}

func (self *Connection) dispatch(msg []byte) {
	self.ircify(msg)
}
