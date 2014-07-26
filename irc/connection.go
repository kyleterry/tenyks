package irc

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/kyleterry/tenyks/config"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Connection struct {
	Name            string
	Config          config.ConnectionConfig
	currentNick     string
	currentServer   string
	nickIndex       int
	connectAttempts uint
	usingSSL        bool
	socket          net.Conn
	In              <-chan string
	Out             chan<- string
	io              *bufio.ReadWriter
	connected       bool
	MessagesRecved  uint
	MessagesSent    uint
	Registry        *HandlerRegistry
	ConnectWait     chan bool
	LastPong        time.Time
	PongIn          chan bool
}

func NewConn(name string, conf config.ConnectionConfig) *Connection {
	registry := NewHandlerRegistry()
	conn := &Connection{
		Name:            name,
		Config:          conf,
		nickIndex:       0,
		connectAttempts: 0,
		usingSSL:        conf.Ssl,
		socket:          nil,
		io:              nil,
		connected:       false,
		Registry:        registry,
		ConnectWait:     make(chan bool, 1),
		PongIn:          make(chan bool, 1),
	}
	conn.addBaseHandlers()
	return conn
}

// Connect is a goroutine that returns a channel that is true if connected
// successfully and false if not.
// It returns a bool channel that when closed or is passed true means success.
func (self *Connection) Connect() chan bool {
	c := make(chan bool, 1)
	go func() {
		retries := 0
		for {
			if retries > self.Config.Retries {
				log.Error("[%s] Max retries reached.",
					self.Name)
				c <- false
				return
			}
			if self.socket != nil {
				break
			}
			server := fmt.Sprintf(
				"%s:%d",
				self.Config.Host,
				self.Config.Port)
			var socket net.Conn
			var err error
			if self.usingSSL {
				socket, err = tls.Dial("tcp", server, nil)
				if err != nil {
					log.Error("[%s] Connection failed... Retrying.",
						self.Name)
					retries += 1
					time.Sleep(time.Second * time.Duration(retries))
					continue
				}
			} else {
				socket, err = net.Dial("tcp", server)
				if err != nil {
					retries += 1
					continue
				}
			}
			self.socket = socket
			self.io = bufio.NewReadWriter(
				bufio.NewReader(self.socket),
				bufio.NewWriter(self.socket))
			self.In = self.recv()
			self.Out = self.send()
			self.connected = true
			c <- true
		}
		c <- true
	}()
	return c
}

// Disconnect will hangup the connection with IRC and reset channels and other
// important bootstrap attributes back to the defaults.
func (self *Connection) Disconnect() {
	if self.connected {
		log.Debug("[%s] Disconnect called", self.Name)
		close(self.Out)
		self.connected = false
		self.socket.Close()
		self.socket = nil
		self.ConnectWait = make(chan bool, 1)
	}
}

// send will kick off a gorouting that will loop forever. It will recieve data
// on a channel and send that to the IRC socket.
// It will return a string channel when called.
func (self *Connection) send() chan<- string {
	c := make(chan string, 1000)
	// goroutine for sending data to the IRC server
	go func() {
		log.Debug("[%s] Starting send loop", self.Name)
		for {
			select {
			case line, ok := <-c:
				if !ok {
					log.Debug("[%s] Stopping send loop", self.Name)
					return
				}
				self.MessagesSent += 1
				self.write(line)
			}
		}
	}()
	return c
}

// write will recieve a string and write it to the IO buffer. It then flushes
// the buffer which in turn will call write() on the socket.
// It might return an error if something goes wrong.
func (self *Connection) write(line string) error {
	_, wrerr := self.io.WriteString(line + "\r\n")
	if wrerr != nil {
		return wrerr
	}
	flerr := self.io.Flush()
	if flerr != nil {
		return flerr
	}
	return nil
}

// recv will kick off a goroutine that will loop forever. It will recieve data
// from a bufio reader and send that to a string channel. Since sockets can 
// send nil when a disconnect occurs, it has a minor responsibility of calling
// the Disconnect method when that happens.
// It will return a string channel when called.
func (self *Connection) recv() <-chan string {
	c := make(chan string, 1000)
	// goroutine for receiving data from the IRC server
	go func() {
		for {
			rawLine, err := self.io.ReadString('\n')
			if err != nil {
				self.Disconnect()
				close(c)
				break
			}
			if rawLine != "" {
				//TODO strip newlines
				self.MessagesRecved += 1
				c <- rawLine
			}
		}
	}()
	return c
}

// watchdog will occasionally PING the server and hope to hear back a PONG.
// If no PONG is recieved in reasonable amount of time, it's safe to assume
// we have been disconnected.
func (self *Connection) watchdog() {
	for {
		<-time.After(time.Second * 60)
		dispatch("send_ping", self, nil)
		select {
		case <-self.PongIn:
			continue
		case <-time.After(time.Second * 20):
			self.Disconnect()
			break
		}
	}
}

// IsConnected can be called to detemine if a connection is still connected.
// It returns a bool
func (self *Connection) IsConnected() bool {
	return self.connected
}

// GetCurrentNick will return the nick currently being used in the IRC connection
// It returns a string
func (self *Connection) GetCurrentNick() string {
	return self.currentNick
}

func (self *Connection) String() string {
	msg := "Tenyks Connection for " + self.Name + ".\n"
	if self.connected {
		msg += "Connected to " + self.currentServer + "\n"
	} else {
		msg += "Disconnected\n"
	}
	return msg
}
