package mockirc_test

import (
	"bufio"
	"fmt"
	"net"
	"log"
	"github.com/kyleterry/tenyks/mockirc"
)

func ExampleIRCInteraction() {
	var client net.Conn
	var err error
	var wait chan bool
	var msg []byte

	ircServer := mockirc.New("mockirc.tenyks.io", 6661) // servername and port
	// When I recieve "PING mockirc.tenyks.io" on the server, respond back with PONG...
	ircServer.When("PING mockirc.tenyks.io").Respond(":PONG mockirc.tenyks.io")
	ircServer.When("NICK kyle").Respond("... response to NICK")
	wait, err = ircServer.Start()
	if err != nil {
		log.Fatal(err)
	}
	defer ircServer.Stop()
	<-wait // wait for start to fire up channel

	client, err = net.Dial("tcp", "localhost:6661")

	io := bufio.NewReadWriter(
		bufio.NewReader(client),
		bufio.NewWriter(client))

	_, err = io.WriteString("PING mockirc.tenyks.io\r\n")
	if err != nil {
		log.Fatal(err)
	}

	msg, err = io.ReadString('\n')

	if string(msg) != ":PONG mockirc.tenyks.io" {
		log.Fatal("Invalid response")
	}
}
