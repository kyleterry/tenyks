package irc

// IRC RFC2812: http://tools.ietf.org/html/rfc2812
// If you make modifications, please follow the spec guidelines.
// If you find code that conflicts with the spec, please file a bug report on Github.

import (
	"bufio"
	"container/list"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/kyleterry/tenyks/config"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")

type Connection struct {
	// Name of the network (e.g. freenode, efnet, localhost)
	Name string
	// Network configuration
	Config config.ConnectionConfig
	// Set to the current nick in use on the server
	currentNick string
	// Set to the current server connected to in the network.
	currentServer string
	// Current index in the array of nicks
	nickIndex int
	// Whether or not we are connected with SSL (TLS)
	usingSSL bool
	// This is the socket used to communicate with IRC
	socket net.Conn
	// Channel used to capture messages coming in from IRC
	In <-chan string
	// Channel used to send messages to IRC
	Out chan<- string
	// bufio readerwriter instance
	io *bufio.ReadWriter
	// are we currently connected?
	connected bool
	// How many messages we have recieved. Reset when tenyks is restarted.
	MessagesRecved uint
	// How mant messages we have sent. Reset when tenyks is restarted.
	MessagesSent uint
	// Created holds the datetime the connection was created.
	Created time.Time
	// Handler registry. These handle various commands from IRC.
	Registry *HandlerRegistry
	// Channel used to tell things to wait for the connection to succeed before spawning goroutines.
	ConnectWait chan bool
	// Last PONG recieved from the network.
	LastPong time.Time
	// Channel for the connection watchdog.
	PongIn chan bool
	// Number of retry attempts the connection made
	retries int
	// Current channels tenyks is in. Jnerula hates state and I don't care.
	Channels *list.List
	// Yes, I'm sharing memory. Sorry, mom.
	channelMutex    *sync.Mutex
	connectionMutex *sync.Mutex
}

// NewConnection will create a new instance of an irc.Connection.
// It will return *irc.Connection
func NewConnection(name string, conf config.ConnectionConfig) *Connection {
	registry := NewHandlerRegistry()
	conn := &Connection{
		Name:            name,
		Config:          conf,
		usingSSL:        conf.Ssl,
		Registry:        registry,
		ConnectWait:     make(chan bool, 1),
		PongIn:          make(chan bool, 1),
		Created:         time.Now(),
		Channels:        list.New(),
		channelMutex:    &sync.Mutex{},
		connectionMutex: &sync.Mutex{},
	}
	conn.addBaseHandlers()
	return conn
}

// Connect is a goroutine that returns a channel that is true if connected
// successfully and false if not.
// It returns a bool channel that when closed or is passed true means success.
func (conn *Connection) Connect() chan bool {
	c := make(chan bool, 1)
	go func() {
		conn.retries = 0
		var (
			socket net.Conn
			err    error
		)
		for {
			if conn.retries > conn.Config.Retries {
				log.Errorf("[%s] Max retries reached.",
					conn.Name)
				c <- false
				return
			}
			if conn.socket != nil {
				break
			}
			server := fmt.Sprintf("%s:%d", conn.Config.Host, conn.Config.Port)
			if conn.usingSSL {
				socket, err = tls.Dial("tcp", server, nil)
				if err != nil {
					log.Errorf("[%s] Connection failed... Retrying.",
						conn.Name)
					conn.retries += 1
					time.Sleep(time.Second * time.Duration(conn.retries))
					continue
				}
			} else {
				socket, err = net.Dial("tcp", server)
				if err != nil {
					conn.retries += 1
					continue
				}
			}
			conn.socket = socket
			conn.io = bufio.NewReadWriter(
				bufio.NewReader(conn.socket),
				bufio.NewWriter(conn.socket))
			conn.In = conn.recv()
			conn.Out = conn.send()
			conn.connected = true
			c <- true
		}
		c <- true
	}()
	return c
}

// Disconnect will hangup the connection with IRC and reset channels and other
// important bootstrap attributes back to the defaults.
func (conn *Connection) Disconnect() {
	if conn.connected {
		log.Debugf("[%s] Disconnect called", conn.Name)
		close(conn.Out)
		conn.connected = false
		conn.socket.Close()
		conn.socket = nil
		conn.ConnectWait = make(chan bool, 1)
	}
}

// send will kick off a gorouting that will loop forever. It will recieve data
// on a channel and send that to the IRC socket.
// It will return a string channel when called.
func (conn *Connection) send() chan<- string {
	c := make(chan string, 1000)
	// goroutine for sending data to the IRC server
	go func() {
		log.Debugf("[%s] Starting send loop", conn.Name)
		for {
			select {
			case line, ok := <-c:
				if !ok {
					log.Debugf("[%s] Stopping send loop", conn.Name)
					return
				}
				conn.connectionMutex.Lock()
				conn.MessagesSent += 1
				conn.connectionMutex.Unlock()
				conn.write(line)
				if conn.Config.FloodProtection {
					duration, _ := time.ParseDuration("500ms")
					<-time.After(duration)
				}
			}
		}
	}()
	return c
}

// write will recieve a string and write it to the IO buffer. It then flushes
// the buffer which in turn will call write() on the socket.
// It might return an error if something goes wrong.
func (conn *Connection) write(line string) error {
	if len(line) > 510 {
		// IRC RFC 2812 states the max length for messages is 512 INCLUDING cr-lf.
		log.Warningf("[%s] Message is too long. Truncating...", conn.Name)
		line = line[:510] // Silently truncate to 510 chars as per IRC spec
	}
	_, wrerr := conn.io.WriteString(line + "\r\n")
	if wrerr != nil {
		return wrerr
	}
	flerr := conn.io.Flush()
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
func (conn *Connection) recv() <-chan string {
	c := make(chan string, 1000)
	// goroutine for receiving data from the IRC server
	go func() {
		for {
			rawLine, err := conn.io.ReadString('\n')
			if err != nil {
				log.Error(err)
				conn.Disconnect()
				close(c)
				break
			}
			if rawLine != "" {
				//TODO strip newlines
				conn.MessagesRecved += 1
				c <- rawLine
			}
		}
	}()
	return c
}

// watchdog will occasionally PING the server and hope to hear back a PONG.
// If no PONG is recieved in reasonable amount of time, it's safe to assume
// we have been disconnected.
func (conn *Connection) watchdog() {
	for {
		<-time.After(time.Second * 60)
		dispatch("send_ping", conn, nil)
		select {
		case <-conn.PongIn:
			continue
		case <-time.After(time.Second * 20):
			conn.Disconnect()
			break
		}
	}
}

// IsConnected can be called to detemine if a connection is still connected.
// It returns a bool
func (conn *Connection) IsConnected() bool {
	return conn.connected
}

func (conn *Connection) GetRetries() int {
	return conn.retries
}

// GetCurrentNick will return the nick currently being used in the IRC connection
// It returns a string
func (conn *Connection) GetCurrentNick() string {
	return conn.currentNick
}

func (conn *Connection) GetInfo() []string {
	var info []string
	conn.connectionMutex.Lock()
	info = append(info,
		fmt.Sprintf("This connection (%s) has been up since %s", conn.Name,
			conn.Created),
		fmt.Sprintf("It has recieved %d messages", conn.MessagesRecved),
		fmt.Sprintf("It has sent %d messages", conn.MessagesSent))
	conn.connectionMutex.Unlock()
	return info
}

func (conn *Connection) IsInChannel(channel string) bool {
	conn.channelMutex.Lock()
	defer conn.channelMutex.Unlock()
	for e := conn.Channels.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == channel {
			return true
		}
	}
	return false
}

func (conn *Connection) GetChannelElement(channel string) *list.Element {
	conn.channelMutex.Lock()
	defer conn.channelMutex.Unlock()
	for e := conn.Channels.Front(); e != nil; e = e.Next() {
		if e.Value.(string) == channel {
			return e
		}
	}
	return nil
}

func (conn *Connection) JoinChannel(channel string) {
	if conn.IsConnected() {
		if !conn.IsInChannel(channel) {
			conn.channelMutex.Lock()
			defer conn.channelMutex.Unlock()
			conn.Out <- fmt.Sprintf("JOIN %s", channel)
			conn.Channels.PushFront(channel)
		}
	}
}

func (conn *Connection) PartChannel(channel string) {
	if conn.IsConnected() {
		if conn.IsInChannel(channel) {
			conn.Out <- fmt.Sprintf("PART %s", channel)
			if e := conn.GetChannelElement(channel); e != nil {
				conn.channelMutex.Lock()
				defer conn.channelMutex.Unlock()
				conn.Channels.Remove(e)
			}
		}
	}
}

func (conn *Connection) String() string {
	msg := "Tenyks Connection for " + conn.Name + "; "
	if conn.connected {
		msg += "Connected to " + conn.currentServer
	} else {
		msg += "Disconnected"
	}
	return msg
}
