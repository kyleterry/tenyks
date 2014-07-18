package service

import (
	"fmt"
	"github.com/kyleterry/tenyks/irc"
)

func ConnectionReactor(ircconns *map[string]*irc.Connection,
	conn *Connection) {
	conn.Bootstrap(ircconns)
	for {
		select {
		case msg := <-conn.In:
			go dispatch(msg)
		}
	}
}

func dispatch(msg []byte) {
	fmt.Println(string(msg[:]))
	ircify(msg)
}
