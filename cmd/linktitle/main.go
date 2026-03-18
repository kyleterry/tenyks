package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	client "github.com/kyleterry/tenyks/pkg/client"
)

var urlRe = regexp.MustCompile(`https?://\S+`)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

func matchURL(msg client.Message) client.Result {
	url := urlRe.FindString(msg.Payload)
	if url == "" {
		return nil
	}
	// Strip trailing punctuation that is commonly not part of the URL.
	url = strings.TrimRight(url, ".,;:!?)'\"")
	return client.Result{"url": url}
}

func fetchTitle(rawURL string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "tenyks-linktitle/1.0")
	req.Header.Set("Accept", "text/html")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		return "", fmt.Errorf("not HTML (%s)", ct)
	}

	// Read at most 64 KB; the title is always in <head>.
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", err
	}

	return extractTitle(string(body))
}

func extractTitle(body string) (string, error) {
	lower := strings.ToLower(body)

	start := strings.Index(lower, "<title")
	if start < 0 {
		return "", fmt.Errorf("no <title> found")
	}
	// Advance past the closing '>' of the opening tag.
	closeTag := strings.Index(lower[start:], ">")
	if closeTag < 0 {
		return "", fmt.Errorf("malformed <title> tag")
	}
	contentStart := start + closeTag + 1

	end := strings.Index(lower[contentStart:], "</title")
	if end < 0 {
		return "", fmt.Errorf("no </title> found")
	}

	title := strings.TrimSpace(body[contentStart : contentStart+end])
	if title == "" {
		return "", fmt.Errorf("empty title")
	}

	// Collapse whitespace (titles often contain newlines/tabs from HTML formatting).
	title = strings.Join(strings.Fields(title), " ")

	return html.UnescapeString(title), nil
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
		Name:        "linktitle",
		Version:     "1.0",
		Description: "Fetches the page title for URLs posted in chat",
		HelpText:    "post a URL and linktitle will reply with the page title",
		Addr:        *addr,
		TLS: client.TLSConfig{
			CAFile:   *caFile,
			CertFile: *certFile,
			KeyFile:  *keyFile,
		},
		Debug: *debug,
	}

	svc := client.New(cfg)

	svc.Handle(client.MsgHandler{
		MatcherFunc: client.MatcherFunc(matchURL),
		MatchHandler: client.HandlerFunc(func(match client.Result, msg client.Message, com *client.Communication) {
			url := match["url"]
			title, err := fetchTitle(url)
			if err != nil {
				log.Printf("linktitle: %s: %v", url, err)
				return
			}
			com.Send(fmt.Sprintf("[ %s ]", title), msg)
		}),
	})

	log.Print("Starting linktitle service")
	if err := svc.Run(); err != nil {
		log.Fatal(err)
	}
}
