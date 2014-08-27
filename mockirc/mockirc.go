package mockirc

import (
	"bufio"
	"fmt"
	"net"
)

type MockIRC struct {
	Port int
	Socket net.Listener
	ctl chan bool
	running bool
}

func New(port int) *MockIRC {
	irc := &MockIRC{}
	if port == 0 {
		irc.Port = 6661
	} else {
		irc.Port = port
	}
	return irc
}

func (irc *MockIRC) Start() error {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", irc.Port))
	if err != nil {
		return err
	}
	irc.Socket = sock
	irc.ctl = make(chan bool, 1)
	go func () {
		for {
			conn, err := irc.Socket.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			go irc.ConnectionWorker(conn)
		}
	}()
	irc.running = true
	return nil
}

func (irc *MockIRC) Stop() {
	
}

func (irc *MockIRC) ConnectionWorker(conn net.Conn) {
	
}
