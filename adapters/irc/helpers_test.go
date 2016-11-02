package irc

import (
	"testing"
)

func TestIsDirect(t *testing.T) {
	direct_msg := "kyle: you are pretty terrible at coding"
	if !IsDirect(direct_msg, "kyle") {
		t.Error("Expected", true, "got", false)
	}

	not_direct_msg := "this isn't a direct message"
	if IsDirect(not_direct_msg, "kyle") {
		t.Error("Expected", false, "got", true)
	}
}

func TestIsChannel(t *testing.T) {
	if !IsChannel("#test") {
		t.Error("Expected", true, "got", false)
	}

	if IsChannel("test") {
		t.Error("Expected", false, "got", true)
	}
}

//func TestStripNickOnDirect(msg string, nick string) {
func TestStripNickOnDirect(t *testing.T) {
	new_msg := StripNickOnDirect("kyle: this is a message", "kyle")
	if new_msg != "this is a message" {
		t.Error("Expected", "this is a message", "got", new_msg)
	}
}
