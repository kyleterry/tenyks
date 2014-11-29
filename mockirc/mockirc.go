package mockirc

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"
)

type MockIRC struct {
	Port       int
	ServerName string
	Socket     net.Listener
	ctl        chan bool
	running    bool
	events     map[string]*WhenEvent
	io         *bufio.ReadWriter
}

func New(server string, port int) *MockIRC {
	irc := &MockIRC{}
	if port == 0 {
		irc.Port = 6661
	} else {
		irc.Port = port
	}
	irc.ServerName = server
	irc.events = make(map[string]*WhenEvent)
	return irc
}

func (irc *MockIRC) Start() (chan bool, error) {
	wait := make(chan bool, 1)
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", irc.Port))
	if err != nil {
		return nil, err
	}
	irc.Socket = sock
	irc.ctl = make(chan bool, 1)
	go func() {
		wait <- true
		close(wait)
		for {
			conn, err := irc.Socket.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			go irc.connectionWorker(conn)
		}
	}()
	irc.running = true
	return wait, nil
}

func (irc *MockIRC) Stop() error {
	err := irc.Socket.Close()
	if err != nil {
		return err
	}
	<-time.After(time.Second)
	return nil
}

func (irc *MockIRC) connectionWorker(conn net.Conn) {
	irc.io = bufio.NewReadWriter(
		bufio.NewReader(conn),
		bufio.NewWriter(conn))
	defer conn.Close()
	for {
		msg, err := irc.io.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		irc.handleMessage(msg)
	}
}

func (irc *MockIRC) handleMessage(msg string) {
	msg = strings.TrimSuffix(msg, "\r\n")
	var err error
	if val, ok := irc.events[msg]; ok {
		for _, response := range val.responses {
			_, err = irc.io.WriteString(response + "\r\n")
			if err != nil {
				fmt.Println(err)
				return
			}
			err = irc.io.Flush()
			if err != nil {
				fmt.Println(err)
				return
			}
		}
	} else {
		fmt.Printf("Nothing to do for %s\n", msg)
	}
	fmt.Println(msg)
}

func (irc *MockIRC) Send(thing string) {
	irc.io.WriteString(thing + "\r\n")
}

type WhenEvent struct {
	event     string
	responses []string
}

func (irc *MockIRC) When(event string) *WhenEvent {
	when := &WhenEvent{event: event}
	irc.events[event] = when
	return when
}

func (when *WhenEvent) Respond(response string) *WhenEvent {
	when.responses = append(when.responses, response)
	return when
}
