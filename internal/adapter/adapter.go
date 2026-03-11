package adapter

import (
	"context"

	servicepb "github.com/kyleterry/tenyks/internal/service"
)

// HandlerFunc is called when an adapter receives an inbound chat message.
type HandlerFunc func(*servicepb.Message)

type Adapter interface {
	GetName() string
	GetType() AdapterType
	Dial(ctx context.Context) error
	Close(ctx context.Context) error
	SendAsync(ctx context.Context, msg *servicepb.Message) error
	RegisterMessageHandler(HandlerFunc)
}

type AdapterType int

const (
	AdapterTypeIRC AdapterType = iota
)

var AdapterTypeMapping = map[string]AdapterType{
	"irc": AdapterTypeIRC,
}
