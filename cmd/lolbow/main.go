package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"
	"sync"

	client "github.com/kyleterry/tenyks/pkg/client"
)

var prettyColors = []int{4, 5, 7, 9, 10, 11, 12, 13, 14, 15}

type channelState struct {
	lollers  map[string]struct{}
	counter  int
	maxChain int
}

func newChannelState() *channelState {
	return &channelState{
		lollers:  make(map[string]struct{}),
		maxChain: randMaxChain(),
	}
}

func randMaxChain() int {
	return rand.IntN(8) + 2 // [2, 9]
}

func (cs *channelState) reset() {
	cs.lollers = make(map[string]struct{})
	cs.counter = 0
	cs.maxChain = randMaxChain()
}

func createRainbow() string {
	var b strings.Builder
	for _, letter := range []byte{'L', 'O', 'L', 'B', 'O', 'W'} {
		color := prettyColors[rand.IntN(len(prettyColors))]
		fmt.Fprintf(&b, "\x03%02d%c", color, letter)
	}
	return b.String()
}

type lolbow struct {
	mu       sync.Mutex
	channels map[string]*channelState
}

func (l *lolbow) state(channel string) *channelState {
	cs, ok := l.channels[channel]
	if !ok {
		cs = newChannelState()
		l.channels[channel] = cs
	}
	return cs
}

func (l *lolbow) handleLol(_ client.Result, msg client.Message, com *client.Communication) {
	l.mu.Lock()
	cs := l.state(msg.Target)

	rainbow := ""
	if _, seen := cs.lollers[msg.Nick]; !seen {
		cs.lollers[msg.Nick] = struct{}{}
		cs.counter++
		if cs.counter >= cs.maxChain {
			rainbow = createRainbow()
			cs.reset()
		}
	}
	l.mu.Unlock()

	if rainbow != "" {
		com.Send(rainbow, msg)
	}
}

func (l *lolbow) handleMaxChain(_ client.Result, msg client.Message, com *client.Communication) {
	l.mu.Lock()
	cs := l.state(msg.Target)
	reply := fmt.Sprintf("%s: max chain is %d, current chain at %d", msg.Nick, cs.maxChain, cs.counter)
	l.mu.Unlock()

	com.Send(reply, msg)
}

func main() {
	addr := flag.String("addr", "", "tenyks server address (default localhost:50001)")
	caFile := flag.String("ca", "", "path to CA certificate file")
	certFile := flag.String("cert", "", "path to client certificate file")
	keyFile := flag.String("key", "", "path to client private key file")
	debug := flag.Bool("debug", false, "enable debug logging")

	flag.Parse()

	if *caFile == "" || *certFile == "" || *keyFile == "" {
		log.Fatal("-ca, -cert, and -key are required")
	}

	cfg := &client.Config{
		Name:        "lolbow",
		Version:     "1.0",
		Description: "A celebration of what it means to truly laugh",
		HelpText:    "say lol — enough unique lollers in a row triggers a LOLBOW",
		Addr:        *addr,
		TLS: client.TLSConfig{
			CAFile:   *caFile,
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
		Debug: *debug,
	}

	lb := &lolbow{channels: make(map[string]*channelState)}

	svc := client.New(cfg)

	svc.Handle(client.MsgHandler{
		MatcherFunc:  client.NewRegexMatcher(`(?i)^max chain$`),
		MatchHandler: client.HandlerFunc(lb.handleMaxChain),
		HelpText:     "max chain - show the current lol chain threshold and progress",
	})

	svc.Handle(client.MsgHandler{
		MatcherFunc:  client.NewRegexMatcher(`(?i)lol`),
		MatchHandler: client.HandlerFunc(lb.handleLol),
		HelpText:     "lol - contribute to the lol chain; enough unique lollers triggers a LOLBOW",
	})

	log.Print("Starting lolbow service")
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}
