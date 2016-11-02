package irc

import(
	"testing"
)

func TestParseMessage(t *testing.T) {
	msg_string := ":kyle!~kyle@localhost PRIVMSG #tenyks :tenyks: messages are awesome"
	msg := ParseMessage(msg_string)
	if msg == nil {
		t.Error("Expected", Message{}, "got", msg)
	}

	if msg.Command != "PRIVMSG" {
		t.Error("Expected", Message{}, "got", msg)
	}

	if msg.Trail != "tenyks: messages are awesome" {
		t.Error("Expected", "tenyks: messages are awesome", "got", msg)
	}

	if msg.Params[0] != "#tenyks" {
		t.Error("Expected", "#tenyks", "got", msg.Params[0])
	}
}
