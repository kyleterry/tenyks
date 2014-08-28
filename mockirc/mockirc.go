package mockirc

import (
	"bufio"
	"fmt"
	"net"
)

type MockIRC struct {
	Port       int
	ServerName string
	Socket     net.Listener
	ctl        chan bool
	running    bool
	events     map[string]*WhenEvent
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

func (irc *MockIRC) Start() error {
	sock, err := net.Listen("tcp", fmt.Sprintf(":%d", irc.Port))
	if err != nil {
		return err
	}
	irc.Socket = sock
	irc.ctl = make(chan bool, 1)
	go func() {
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
	return nil
}

func (irc *MockIRC) Stop() {
	irc.Socket.Close()
}

func (irc *MockIRC) connectionWorker(conn net.Conn) {
	io := bufio.NewReadWriter(
		bufio.NewReader(conn),
		bufio.NewWriter(conn))
	defer conn.Close()
	for {
		msg, err := io.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		}
		irc.handleMessage(msg, conn)
	}
}

func (irc *MockIRC) handleMessage(msg string, conn net.Conn) {
	fmt.Println(msg)
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
