package irc

import (
	"testing"
	"time"
	"strings"
	"fmt"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/mockirc"
	_"code.google.com/p/gomock/gomock"
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
		Port: 6661,
		FloodProtection: true,
		Retries: 5,
		Nicks: []string{"tenyks", "tenyks-"},
		Ident: "something",
		Realname: "tenyks",
		Admins: []string{"kyle"},
		Channels: []string{"#tenyks", "#test"},
	}
}

func TestCanConnectToIRC(t *testing.T) {
	ircServer := mockirc.New("mockirc.tenyks.io", 0)
	ircServer.When("PING mockirc.tenyks.io").Respond(":PONG mockirc.tenyks.io")
	ircServer.Start()
	defer ircServer.Stop()

	conn := NewConn("mockirc", MakeConnConfig())
	wait := conn.Connect()
	<-wait
	defer conn.Disconnect()
	
	if conn.GetRetries() > 0 {
		t.Error("Expected", 0, "got", conn.GetRetries())
	}

	if !conn.IsConnected() {
		t.Error("Expected", true, "got", false)
	}

	dispatch("bootstrap", conn, nil)
}
