package main

import (
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"time"

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
		panic(err)
	}

	args := &control.ChannelArgs{"localhost", "#test"}

	err = client.Call("ControlServer.JoinChannel", args, &reply)
	if err != nil {
		panic(err)
	}

	<-time.After(time.Second * 10)

	err = client.Call("ControlServer.PartChannel", args, &reply)
	if err != nil {
		panic(err)
	}

	log.Println("Made RPC call")
}
