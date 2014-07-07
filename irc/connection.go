package irc

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/kyleterry/tenyks/config"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Connection struct {
	Name            string
	Config          config.ConnectionConfig
	connectAttempts uint
	usingSSL        bool
	socket          net.Conn
	In              <-chan string
	Out             chan<- string
	sendCtl         chan bool
	io              *bufio.ReadWriter
	connected       bool
	MessagesRecved	uint
	MessagesSent	uint
}

func NewConn(name string, conf config.ConnectionConfig) *Connection {
	sendCtl := make(chan bool, 1)
	return &Connection{
		Name:            name,
		Config:          conf,
		connectAttempts: 0,
		usingSSL:        conf.Ssl,
		socket:          nil,
		sendCtl:         sendCtl,
		io:              nil,
		connected:       false,
	}
}

func Bootstrap(conn *Connection) {
	if conn.IsConnected() {
		conn.Out <- fmt.Sprintf(
			"USER %s %s %s :%s",
			conn.Config.Nicks[0],
			conn.Config.Host,
			conn.Config.Ident,
			conn.Config.Realname)
		conn.Out <- fmt.Sprintf(
			"NICK %s", conn.Config.Nicks[0])
		for _, channel := range conn.Config.Channels {
			conn.Out <- fmt.Sprintf("JOIN %s", channel)
		}
	}
}

// Goroutine that returns a channel that is true if connected successfully
// and false if not.
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

func (self *Connection) Disconnect() {
	self.sendCtl <- true
	close(self.Out)
	self.connected = false
	self.socket.Close()
	self.socket = nil
}

func (self *Connection) send() chan<- string {
	c := make(chan string, 1000)
	// goroutine for sending data to the IRC server
	go func() {
		log.Debug("[%s] Starting send loop", self.Name)
		for {
			select {
			case line := <-c:
				log.Debug("[%s] Sending %s", self.Name, line)
				self.MessagesSent += 1
				self.write(line)
			case <-self.sendCtl:
				log.Debug("[%s] Stopping send loop", self.Name)
				return
			}
		}
	}()
	return c
}

func (self *Connection) write(line string) error {
	self.io.WriteString(line + "\r\n")
	self.io.Flush()
	return nil
}

// Recv channel factory
func (self *Connection) recv() <-chan string {
	c := make(chan string, 1000)
	// goroutine for receiving data from the IRC server
	go func() {
		for {
			rawLine, err := self.io.ReadString('\n')
			if err != nil {
				self.Disconnect()
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

func (self *Connection) IsConnected() bool {
	return self.connected
}

func (self *Connection) String() string {
	msg := "Tenyks Connection to " + self.Name + ".\n"
	if self.connected {
		msg += "Connected\n"
	} else {
		msg += "Disconnected\n"
	}
	return msg
}
