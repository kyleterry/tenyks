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
