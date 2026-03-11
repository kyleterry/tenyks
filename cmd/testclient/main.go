package main

import (
	"context"
	"flag"
	"log"

	"github.com/kyleterry/tenyks/internal/client"
	"github.com/kyleterry/tenyks/internal/config"
	"github.com/kyleterry/tenyks/internal/tlsconfig"
)

func main() {
	addr := flag.String("addr", client.DefaultAddr, "tenyks server address")
	caFile := flag.String("ca", "", "path to CA certificate file")
	certFile := flag.String("cert", "", "path to client certificate file")
	keyFile := flag.String("key", "", "path to client private key file")

	flag.Parse()

	if *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatal("-ca, -cert, and -key are required")
	}

	certs, err := tlsconfig.Load(&config.TLS{
		CAFile:         *caFile,
		CertFile:       *certFile,
		PrivateKeyFile: *keyFile,
	})
	if err != nil {
		log.Fatal(err)
	}

	c := client.NewClient(*addr, tlsconfig.NewClientConfig(certs))

	if err := c.Dial(context.Background()); err != nil {
		log.Fatal(err)
	}

	defer c.Close()

	log.Println("connected to", *addr)
}
