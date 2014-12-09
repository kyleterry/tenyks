package irc

import (
	"testing"
	"time"
)

func ManagerCanRespondTest(t *testing.T) {
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

	rcom <- request
	cmcom <- rcom

	select {
	case request = <-rcom:
	case <-time.After(time.Second * 1):
		t.Error("Timed out and while waiting for a response")
	}
}
