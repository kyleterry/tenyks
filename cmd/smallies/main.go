package main

import (
	"flag"
	"log"

	"github.com/kyleterry/tenyks/pkg/client"
)

func main() {
	addr := flag.String("addr", "", "tenyks server address (default localhost:50001)")
	caFile := flag.String("ca", "", "path to CA certificate file")
	certFile := flag.String("cert", "", "path to client certificate file")
	keyFile := flag.String("key", "", "path to client private key file")
	scriptsDir := flag.String("scripts", "scripts", "directory containing .scm script files")
	debug := flag.Bool("debug", false, "enable debug logging")

	flag.Parse()

	if *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatal("-ca, -cert, and -key are required")
	}

	eng := newEngine()
	if err := eng.loadDir(*scriptsDir); err != nil {
		log.Fatalf("smallies: %v", err)
	}

	cfg := &client.Config{
		Name:        "smallies",
		Version:     "1.0",
		Description: "Runs scripts that handle small tasks for incoming messages. Useful for avoiding construction of service clients for small things.",
		HelpText:    "Write scripts in glerp scheme to handle patterns in messages",
		Addr:        *addr,
		TLS: client.TLSConfig{
			CAFile:   *caFile,
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
		Debug: *debug,
	}

	svc := client.New(cfg)
	svc.DefaultHandle(client.MsgHandler{
		MatchHandler: client.HandlerFunc(func(_ client.Result, msg client.Message, com *client.Communication) {
			for _, reply := range eng.dispatch(msg.Nick, msg.Target, msg.Payload) {
				com.Send(reply, msg)
			}
		}),
	})

	log.Print("Starting smallies service")
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}
