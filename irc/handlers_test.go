package irc

import (
	"container/list"
	"testing"
)

func TestHandler(t *testing.T) {
	registry := NewHandlerRegistry()
	if len(registry.Handlers) > 0 {
		t.Error("Expected", 0, "Got", len(registry.Handlers))
	}
	
	my_strings := list.New()
	updater := func(l *list.List) { l.PushBack("updated") }
	handler := NewHandler(func(p ...interface{}) {
		updater(p[0].(*list.List))
	})
	registry.AddHandler("testing", handler)

	if len(registry.Handlers) == 0 {
		t.Error("Expected", 1, "got", len(registry.Handlers))
	}

	i := registry.Handlers["testing"].Front()
	h := i.Value.(*Handler)
	h.Fn(my_strings)
	if v := my_strings.Front().Value.(string); v != "updated" {
		t.Error("Expected", "updated", "got", v)
	}
}
