package irc

import (
	"testing"
	"time"
	"strings"
	"fmt"
	"github.com/kyleterry/tenyks/config"
	_"code.google.com/p/gomock/gomock"
)

func TestNewConn(t *testing.T) {
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
