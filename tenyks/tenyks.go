package main

import (
	"fmt"
	stdlog "log"
	"os"
	"flag"

	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/control"
	"github.com/kyleterry/tenyks/irc"
	"github.com/kyleterry/tenyks/service"
	. "github.com/kyleterry/tenyks/version"
	"github.com/op/go-logging"
)

const (
	Usage = `
Usage: %s [-config <CONFIG PATH>] [OPTIONS]
	CONFIG PATH:
		Path to a json configuration. If none is specified, Tenyks will look
		for a config in common paths (e.g. /etc/tenyks/config.json)
	
	OPTIONS:
		-version, -V
			Used to print Tenyks' version number

		-help, -h
			This help
`
	banner = `
  _                   _         
 | |                 | |        
 | |_ ___ _ __  _   _| | _____  
 | __/ _ \ '_ \| | | | |/ / __| 
 | ||  __/ | | | |_| |   <\__ \ The IRC bot for hackers.
  \__\___|_| |_|\__, |_|\_\___/ 
                 __/ |          
                |___/           
`
)

var (
	log = logging.MustGetLogger("tenyks")
	configPath = flag.String("config", "", "Path to a configuration file")
	versionFlag = flag.Bool("version", false, "Get the current version")
	helpFlag = flag.Bool("help", false, "Get some help")
)

func init() {
	flag.BoolVar(versionFlag, "v", false, "Get the current version")
	flag.BoolVar(helpFlag, "h", false, "Get some help")
}

func main() {

	flag.Parse()

	if *versionFlag {
		fmt.Println("Tenyks version " + TenyksVersion)
		os.Exit(0)
	}

	if *helpFlag {
		fmt.Printf(Usage, os.Args[0])
		os.Exit(0)
	}

	quit := make(chan bool, 1)

	fmt.Printf(banner + "\n")
	fmt.Printf(" Version: %s\n\n", TenyksVersion)

	config.ConfigSearch.AddPath("/etc/tenyks/config.json")
	config.ConfigSearch.AddPath(os.Getenv("HOME") + "/.config/tenyks/config.json")

	// Make configuration from json file
	conf, conferr := config.NewConfigAutoDiscover(configPath)
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

	// Starting Control Server
	var (
		err error
		wait chan bool
		controlServer *control.ControlServer
	)
	if conf.Control.Enabled {
		log.Debug("Control Server is On")
		controlServer, err = control.NewControlServer(conf.Control)
		wait, err = controlServer.Start()
		if err != nil {
			log.Fatal(err)
		}
		<-wait
		log.Info("Control server listening on %s", conf.Control.Bind)
	} else {
		log.Debug("Control Server is Off")
	}

	// Connections map
	connections := make(irc.IRCConnections)
	ircReactors := make([]<-chan bool, 0)

	eng := service.NewServiceEngine(conf.Redis)

	// Create connection, spawn reactors and add to the map
	for _, c := range conf.Connections {
		conn := irc.NewConnection(c.Name, c)
		ctl := make(<-chan bool, 1)
		ircReactors = append(ircReactors, ctl)
		go irc.ConnectionReactor(conn, ctl)
		connections[c.Name] = conn
	}

	eng.SetIRCConns(connections)
	if conf.Control.Enabled {
		controlServer.SetIRCConns(connections)
	}

	for _, ircconn := range connections {
		go eng.RegisterIrcHandlersFor(ircconn)
	}
	go eng.Start()


	<-quit
}
