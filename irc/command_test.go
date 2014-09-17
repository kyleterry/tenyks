package irc

import (
	"testing"
)

func TestSlashCommandConversion(t *testing.T) {
	initCommandHandlers()

	cmd := "/msg foobar sup?!"
	expected := "PRIVMSG foobar :sup?!"

	got, err1 := ConvertSlashCommand(cmd)

	if got != expected {
		t.Error("Expected", expected, "got", got)
	} else if err1 != nil {
		t.Error("Error", err1)
	}

	cmd = "/notacommand foobar"
	_, err2 := ConvertSlashCommand(cmd)
	if err2 == nil {
		t.Error("Expected", "error", "got", nil)
	}
}

func TestJoinCommandConversion(t *testing.T) {
	initCommandHandlers()

	cmd := "/join #foobar"
	expected := "JOIN #foobar"

	got, err := ConvertSlashCommand(cmd)

	if got != expected {
		t.Error("Expected", expected, "got", got)
	} else if err != nil {
		t.Error("Error", err)
	}
}
