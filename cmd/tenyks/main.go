package main

import (
	"context"
	"flag"
	"log"
	"net"

	"github.com/kyleterry/tenyks/internal/adapter"
	"github.com/kyleterry/tenyks/internal/adapter/irc"
	"github.com/kyleterry/tenyks/internal/config"
	"github.com/kyleterry/tenyks/internal/logger"
	"github.com/kyleterry/tenyks/internal/service"
	"github.com/kyleterry/tenyks/internal/tlsconfig"
)

func main() {
	configPath := flag.String("config", "", "path to a configuration file")

	flag.Parse()

	cfg, err := config.NewFromFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	loggerConf := logger.StandardLoggerConfig{
		Debug:         cfg.Logging.Debug,
		ShowTimestamp: true,
	}

	standardLogger := logger.NewStandardLogger("tenyks", loggerConf)

	adapterRegistry := adapter.NewRegistry()

	for _, sc := range cfg.Servers {
		at, ok := adapter.AdapterTypeMapping[sc.Kind]
		if !ok {
			log.Fatalf("no such adapter %s", sc.Kind)
		}

		switch at {
		case adapter.AdapterTypeIRC:
			ircConfig := sc.Config.(config.IRCServer)

			c, err := irc.New(irc.Config{
				Name:     ircConfig.Name,
				Server:   ircConfig.ServerAddr,
				UseTLS:   ircConfig.UseTLS,
				Password: ircConfig.Password,
				User:     ircConfig.User,
				RealName: ircConfig.RealName,
				Nicks:    ircConfig.Nicks,
				Channels: ircConfig.Channels,
				Commands: ircConfig.Commands,
				Logger:   standardLogger,
			})
			if err != nil {
				log.Fatal(err)
			}

			if err := c.Dial(context.Background()); err != nil {
				log.Fatal(err)
			}

			adapterRegistry.RegisterAdapter(c)
		}
	}

	bindAddr := ":50001"
	if cfg.Service != nil && cfg.Service.BindAddr != "" {
		bindAddr = cfg.Service.BindAddr
	}

	l, err := net.Listen("tcp", bindAddr)
	if err != nil {
		log.Fatal(err)
	}

	certs, err := tlsconfig.Load(cfg.Service.TLS)
	if err != nil {
		log.Fatal(err)
	}

	svcServer, gs := service.New(tlsconfig.NewServerConfig(certs), adapterRegistry)

	// Register the service server as a handler on every adapter so that
	// incoming IRC messages are broadcast to all connected service clients.
	for _, a := range adapterRegistry.GetAdaptersFor(adapter.AdapterTypeIRC) {
		a.RegisterMessageHandler(svcServer.Broadcast)
	}

	if err := gs.Serve(l); err != nil {
		log.Fatal(err)
	}
}
