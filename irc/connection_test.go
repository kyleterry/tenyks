package irc

import (
	"testing"
	"time"
	"strings"
	"fmt"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/mockirc"
)

func TestNewConnNoDial(t *testing.T) {
	conf := config.ConnectionConfig{
		Name: "test",
		Ssl: true,
	}
	conn := NewConn(conf.Name, conf)
	if conn.Name != conf.Name {
		t.Error("Expected %s, got %s", conn.Name, conf.Name)
	}

	if !conn.usingSSL {
		t.Error("SSL is supposed to be enabled")
	}

	if conn.IsConnected() {
		t.Error("Connection is not supposed to be connected")
	}

	strMethodResult := fmt.Sprintf("%s", conn)
	if !strings.Contains(strMethodResult, "Disconnected") {
		t.Error("String method seems to be broken, Expected to contain 'Disconnected', got ", strMethodResult)
	}

	select {
	case <-conn.ConnectWait:
		t.Error("Channel is supposed to remain open and not recieve")
	case <-time.After(time.Second):
		break
	}
}

func MakeConnConfig() config.ConnectionConfig {
	return config.ConnectionConfig{
		Name: "mockirc",
		Host: "localhost",
		Port: 26661,
		FloodProtection: true,
		Retries: 5,
		Nicks: []string{"tenyks", "tenyks-"},
		Ident: "something",
		Realname: "tenyks",
		Admins: []string{"kyle"},
		Channels: []string{"#tenyks", "#test"},
	}
}

func TestCanConnectAndDisconnect(t *testing.T) {
	var wait chan bool
	var err error
	ircServer := mockirc.New("mockirc.tenyks.io", 26661)
	wait, err = ircServer.Start()
	if err != nil {
		t.Fatal("Expected nil", "got", err)
	}
	<-wait
	defer ircServer.Stop()

	conn := NewConn("mockirc", MakeConnConfig())
	wait = conn.Connect()
	<-wait
	
	if conn.GetRetries() > 0 {
		t.Error("Expected", 0, "got", conn.GetRetries())
	}

	if !conn.IsConnected() {
		t.Error("Expected", true, "got", false)
	}

	conn.Disconnect()

	if conn.IsConnected() {
		t.Error("Expected", false, "got", true)
	}
}

func TestCanHandshakeAndWorkWithIRC(t *testing.T) {
	var wait chan bool
	var err error
	ircServer := mockirc.New("mockirc.tenyks.io", 26661)
	ircServer.When("USER tenyks localhost something :tenyks").Respond(":101 :Welcome")
	ircServer.When("PING ").Respond(":PONG")
	wait, err = ircServer.Start()
	if err != nil {
		t.Fatal("Expected nil", "got", err)
	}
	defer ircServer.Stop()
	<-wait

	conn := NewConn("mockirc", MakeConnConfig())
	wait = conn.Connect()
	<-wait

	conn.BootstrapHandler(nil)
	<-conn.ConnectWait

	msg := string(<-conn.In)
	if msg != ":101 :Welcome\r\n" {
		t.Error("Expected :101 :Welcome", "got", msg)
	}

	conn.SendPing(nil)
	select {
	case msg := <-conn.In:
		if msg != ":PONG\r\n" {
			t.Error("Expected", ":PONG mockirc", "got", msg)
		}
	case <-time.After(time.Second * 5):
		t.Error("Timed out before getting a response back from mockirc")
	}

	conn.Disconnect()
}
