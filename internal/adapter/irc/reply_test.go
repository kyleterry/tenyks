package irc

import (
	"reflect"
	"testing"
)

func TestNamesReply(t *testing.T) {
	raw := ":user!~nick@server/mask 353 nick = #tenyks :nick1 nick2 nick3 nick4"

	msg, err := NewRawMessageDecoder().Decode(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	nr := NamesReply{m: msg}

	if err := nr.Validate(); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}

	if got, want := nr.Names(), []string{"nick1", "nick2", "nick3", "nick4"}; !reflect.DeepEqual(got, want) {
		t.Errorf("Names() = %v, want %v", got, want)
	}
	if got, want := nr.Channel(), "#tenyks"; got != want {
		t.Errorf("Channel() = %q, want %q", got, want)
	}
}
