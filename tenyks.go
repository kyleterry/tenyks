package main

import (
	"fmt"
	stdlog "log"
	"os"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/kyleterry/tenyks/service"
	. "github.com/kyleterry/tenyks/version"
	"github.com/op/go-logging"
)

const (
	Usage = `
Usage: %s [config path | options]
	Config path:
		Path to a json configuration. If none is specified, Tenyks will look
		for a config in common paths (e.g. /etc/tenyks/config.json)
	
	Options:
		--version, -V
			Used to print Tenyks' version number

		--help, -h
			This help
`
)

var log = logging.MustGetLogger("tenyks")
var connections irc.IrcConnections
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
	// Check for version flag if len(os.Args) > 1 {
	if len(os.Args) > 1 {
		if os.Args[1] == "--version" || os.Args[1] == "-V" {
			fmt.Println("Tenyks version " + TenyksVersion)
			os.Exit(0)
		} else if os.Args[1][0] == '-' {
			fmt.Printf(Usage, os.Args[0])
			os.Exit(0)
		}
	}

	quit := make(chan bool, 1)

	fmt.Printf(banner + "\n")
	fmt.Printf(" Version: %s\n\n", TenyksVersion)

	// Make configuration from json file
	conf, conferr := config.NewConfigAutoDiscover()
	if conferr != nil {
		log.Fatal(conferr)
	}
	conf.Version = TenyksVersion

	// Configure logging
	switch conf.LogLocation {
	case "syslog":
		logBackend, logErr := logging.NewSyslogBackend("")
		if logErr != nil {
			log.Fatal(logErr)
		}
		logging.SetBackend(logBackend)
	default:
	case "stdout":
		flags := stdlog.LstdFlags
		if conf.Debug {
			flags = flags|stdlog.Lshortfile
		}
		logBackend := logging.NewLogBackend(os.Stdout, "", flags)
		logBackend.Color = true
		logging.SetBackend(logBackend)
	}
	if conf.Debug {
		logging.SetLevel(logging.DEBUG, "tenyks")
	} else {
		logging.SetLevel(logging.INFO, "tenyks")
	}

	eng := service.NewServiceEngine(conf.Redis, connections)

	// Connections map
	connections = make(irc.IrcConnections)
	ircReactors = make([]<-chan bool, 0)

	// Create connection, spawn reactors and add to the map
	for _, c := range conf.Connections {
		conn := irc.NewConn(c.Name, c)
		ctl := make(<-chan bool, 1)
		ircReactors = append(ircReactors, ctl)
		go irc.ConnectionReactor(conn, ctl)
		connections[c.Name] = conn
	}

	for _, ircconn := range connections {
		go eng.RegisterIrcHandlersFor(ircconn)
	}
	go eng.Start()

	<-quit
}
