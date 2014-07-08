package main

import (
	stdlog "log"
	"os"
	"fmt"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/kyleterry/tenyks/service"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("tenyks")
var connections map[string]*irc.Connection
var ircReactors []<-chan bool

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
	ircReactors = make([]<-chan bool, 0)

	// Create connection, spawn reactors and add to the map
	for _, c := range conf.Connections {
		conn := irc.NewConn(c.Name, c)
		ctl := make(<-chan bool, 1)
		ircReactors = append(ircReactors, ctl)
		go irc.ConnectionReactor(conn, ctl)
		connections[c.Name] = conn
	}

	serviceConnection := service.NewConn(conf.Redis)
	go service.ConnectionReactor(serviceConnection)

	<-quit
}
