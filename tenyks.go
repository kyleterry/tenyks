package main

import (
	stdlog "log"
	"os"
	"fmt"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")
var connections map[string]*irc.Connection
var ircObservers []<-chan bool
var ircDispatchCommands map[string]bool = make(map[string]bool)
var serviceDispatchCommands map[string]bool = make(map[string]bool)

var banner string = `
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \ 
  \__\___|_| |_|\__, |_|\_\___/ 
                 __/ |          
                |___/           
`

func connectionObserver(conn *irc.Connection, observerCtl <-chan bool) {
	log.Info("[%s] Connecting", conn.Name)
	connected := <-conn.Connect()
	if connected == true {
		irc.Bootstrap(conn)
		for {
			select {
			case rawmsg := <-conn.In:
				msg := irc.ParseMessage(rawmsg)
				log.Debug("%+v", *msg)
				if msg != nil { // Just ignore invalid messages. Who knows...
					dispatch(conn, msg)
				}
			case <-observerCtl:
				break
			}
		}
	} else {
		log.Error("[%s] Could not connect.", conn.Name)
	}
}

func dispatch(conn *irc.Connection, msg *irc.Message) {
	if ircDispatchCommands[msg.Command] {
		if msg.Command == "PING" {
			conn.Out <- fmt.Sprintf("PONG %s\r\n", msg.Trail)
		}

	} else if serviceDispatchCommands[msg.Command] {

	}
}

func main() {
	quit := make(chan bool, 1)

	fmt.Printf(banner + "\n")

	// Make configuration from json file
	conf, conferr := config.NewConfigAutoDiscover()
	if conferr != nil {
		log.Fatal(conferr)
	}

	// Configure logging
	logBackend := logging.NewLogBackend(os.Stdout, "", stdlog.LstdFlags|stdlog.Lshortfile)
	logBackend.Color = true
	logging.SetBackend(logBackend)
	if conf.Debug {
		logging.SetLevel(logging.DEBUG, "tenyks")
	} else {
		logging.SetLevel(logging.INFO, "tenyks")
	}

	// Connections map
	connections = make(map[string]*irc.Connection)
	ircObservers = make([]<-chan bool, 0)

	ircDispatchCommands["PING"] = true
	serviceDispatchCommands["PRIVMSG"] = true

	// Create connection, spawn observers and add to the map
	for _, c := range conf.Connections {
		conn := irc.NewConn(c.Name, c)
		ctl := make(<-chan bool, 1)
		ircObservers = append(ircObservers, ctl)
		go connectionObserver(conn, ctl)
		connections[c.Name] = conn
	}

	<-quit
}
