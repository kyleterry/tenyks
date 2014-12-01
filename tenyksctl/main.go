package main

import(
	"net/rpc"
	"flag"
	"fmt"
	"log"
	
	"github.com/kyleterry/tenyks/control"
)

var (
	hostFlag = flag.String("host", "127.0.0.1", "Host to connect to")
	portFlag = flag.Int("port", 12666, "Port to connect to")
)

func init() {
	flag.StringVar(hostFlag, "h", "127.0.0.1", "Host to connect to")
	flag.IntVar(portFlag, "p", 12666, "Port to connect to")
}

func main() {
	flag.Parse()

	var reply string
	var err error
	var client *rpc.Client

	client, err = rpc.Dial("tcp", fmt.Sprintf("%s:%d", *hostFlag, *portFlag))
	if err != nil {
		log.Fatal(err)
	}

	args := &control.ChannelArgs{"vagrant", "#test"}

	err = client.Call("ControlServer.JoinChannel", args, &reply)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Made RPC call")
}
