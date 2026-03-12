package client

import (
	pb "github.com/kyleterry/tenyks/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler handles a matched message.
type Handler interface {
	HandleMatch(Result, Message, *Communication)
}

// HandlerFunc wraps a plain function as a Handler.
type HandlerFunc func(Result, Message, *Communication)

func (fn HandlerFunc) HandleMatch(r Result, m Message, c *Communication) {
	fn(r, m, c)
}

// MsgHandler pairs a Matcher with the Handler to call when it matches.
type MsgHandler struct {
	// MatcherFunc is used to test incoming messages. Nil means always match.
	MatcherFunc Matcher
	// DirectOnly restricts this handler to messages directed at the bot
	// (DMs or channel mentions).
	DirectOnly bool
	// PrivateOnly restricts this handler to direct messages only (not channels).
	PrivateOnly bool
	// MatchHandler is called when MatcherFunc returns a non-nil Result.
	MatchHandler Handler
	// HelpText documents this handler for help queries.
	HelpText string
}

// NoopHandler is a handler that does nothing. Used as the default handler.
var NoopHandler = MsgHandler{
	MatchHandler: HandlerFunc(func(_ Result, _ Message, _ *Communication) {}),
}

// Communication lets a handler send a reply back through tenyks.
type Communication struct {
	stream pbStream
}

// Send sends line as a PRIVMSG reply. The destination mirrors the original
// message: channel messages reply to the channel, DMs reply to the sender.
func (c *Communication) Send(line string, msg Message) {
	dest := msg.Target
	if !msg.FromChannel {
		dest = msg.Nick
	}

	_ = c.stream.Send(&pb.Message{
		Payload: &pb.Message_Chat{
			Chat: &pb.Chat{
				DestinationPath: dest,
				OriginPath:      msg.Target,
			},
		},
		Content:   line,
		CreatedAt: timestamppb.Now(),
	})
}
