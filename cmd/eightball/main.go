package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"

	client "github.com/kyleterry/tenyks/pkg/client"
)

var responses = []string{
	"It is certain.",
	"It is decidedly so.",
	"Without a doubt.",
	"Yes, definitely.",
	"You may rely on it.",
	"As I see it, yes.",
	"Most likely.",
	"Outlook good.",
	"Yes.",
	"Signs point to yes.",
	"Reply hazy, try again.",
	"Ask again later.",
	"Better not tell you now.",
	"Cannot predict now.",
	"Concentrate and ask again.",
	"Don't count on it.",
	"My reply is no.",
	"My sources say no.",
	"Outlook not so good.",
	"Very doubtful.",
}

var match8ball = client.NewRegexMatcher(`^(?:\S+: )?8ball$`)

func main() {
	addr := flag.String("addr", "", "tenyks server address (default localhost:50001)")
	caFile := flag.String("ca", "", "path to CA certificate file")
	certFile := flag.String("cert", "", "path to client certificate file")
	keyFile := flag.String("key", "", "path to client private key file")

	flag.Parse()

	if *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatal("-ca, -cert, and -key are required")
	}

	cfg := &client.Config{
		Name:        "eightball",
		Version:     "1.0",
		Description: "Magic 8-ball: ask a yes/no question and receive wisdom",
		HelpText:    "<nick>: 8ball - the 8-ball will reveal your fate",
		Addr:        *addr,
		TLS: client.TLSConfig{
			CAFile:   *caFile,
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
	}

	svc := client.New(cfg)

	svc.Handle(client.MsgHandler{
		MatcherFunc: match8ball,
		DirectOnly:  true,
		MatchHandler: client.HandlerFunc(func(_ client.Result, msg client.Message, com *client.Communication) {
			response := responses[rand.IntN(len(responses))]
			com.Send(fmt.Sprintf("%s: \U0001f3b1 %s", msg.Nick, response), msg)
		}),
		HelpText: "<nick>: 8ball - ask the magic 8-ball a yes/no question",
	})

	log.Print("Starting eightball service")
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}
