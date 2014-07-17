package service

import (
	"fmt"
)

func ConnectionReactor(conn *Connection) {
	conn.Bootstrap()
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
