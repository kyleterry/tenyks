package irc

import (
	"testing"
	"time"
	"fmt"

	"github.com/kyleterry/tenyks/mockirc"
)

func TestCanStartCm(t *testing.T) {
	done := make(chan bool)
	cm := NewConnectionManager()
	cmcom := cm.Start(done)

	c := make(chan Request)
	cmcom <- c

	close(done)
}

func TestManagerCanRespond(t *testing.T) {
	done := make(chan bool)
	cm := NewConnectionManager()
	cmcom := cm.Start(done)
	rcom := make(chan Request)
	var request Request

	request = Request{
		RequestType: RtNewConn,
		Requirement: RqResponseRequired,
		Payload: NewConnection("mockirc", MakeConnConfig()),
	}

	cmcom <- rcom
	rcom <- request

	select {
	case request = <-rcom:
	case <-time.After(time.Second * 1):
		t.Error("Timed out and while waiting for a response")
	}

	close(done)
}

func TestManagerCanConnectAndDisconnect(t *testing.T) {
	done := make(chan bool)
	cm := NewConnectionManager()
	cmcom := cm.Start(done)
	rcom := make(chan Request)
	var(
		request Request
		wait chan bool
		err error
	)
	ircServer := mockirc.New("mockirc.tenyks.io", 26661)
	ircServer.When("USER tenyks localhost something :tenyks").Respond(":101 :Welcome")
	ircServer.When("PING ").Respond(":PONG")
	wait, err = ircServer.Start()
	if err != nil {
		t.Fatal("Expected nil", "got", err)
	}
	<-wait

	config := MakeConnConfig()
	conn := NewConnection("mockirc", config)

	fmt.Println("Building request")

	request = Request{
		RequestType: RtNewConn,
		Requirement: RqNoResponse,
		Payload: conn,
	}

	cmcom <- rcom
	rcom <- request

	fmt.Println("Sending request on rcom channel")

	fmt.Println("Waiting for connect")
	var i int
	for i = 0; i <= 4; i++ {
		<-time.After(time.Second * 1)
		if !conn.IsConnected() && i == 4 {
			t.Error("The connection did not connect")
			return
		} else if conn.IsConnected() {
			break
		}
	}
	fmt.Println("connected!")

	request = Request {
		RequestType: RtDisconnect,
		Requirement: RqNoResponse,
		Payload: "mockirc",
	}

	rcom = make(chan Request)
	cmcom <- rcom
	rcom <- request

	fmt.Println("Waiting for disconnect")
	for i = 0; i <= 4; i++ {
		<-time.After(time.Second * 1)
		if conn.IsConnected() && i == 4 {
			t.Error("The connection did not disconnect")
			return
		} else if !conn.IsConnected() {
			break
		}
	}
	fmt.Println("disconnected!")

	err = ircServer.Stop()
	if err != nil {
		t.Fatal("Error stopping mockirc server")
	}
}
