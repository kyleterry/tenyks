package service

import (
	"fmt"
)

func ConnectionReactor(conn *Connection) {
	conn.Bootstrap()
	for {
		select {
		case msg := <-conn.In:
			fmt.Println(string(msg[:]))
		}
	}
}
