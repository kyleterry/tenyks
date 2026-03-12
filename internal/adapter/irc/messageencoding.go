package irc

import (
	"errors"
	"fmt"
	"path"
	"strings"

	servicepb "github.com/kyleterry/tenyks/internal/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MessageEncoder takes a Message object and returns a string value for that
// message
type MessageEncoder interface {
	Encode(*Message) (string, error)
}

// MessageDecoder takes a message as a string and parses it into a Message object
type MessageDecoder interface {
	Decode(string) (*Message, error)
}

// RawMessageEncoder builds a raw IRC message string from a Message object
type RawMessageEncoder struct{}

func (e *RawMessageEncoder) Encode(msg *Message) (string, error) {
	params := msg.Params

	if msg.Trail != "" {
		params = append(params, fmt.Sprintf(":%s", msg.Trail))
	}

	msg.RawMsg = fmt.Sprintf("%s %s", msg.Command, strings.Join(params, " "))

	// TODO encode tags
	return fmt.Sprintf("%s\r\n", msg.RawMsg), nil
}

// NewRawMessageEncoder returns a new RawMessageEncoder object. It is used to
// encode Message objects into strings that can be sent to IRC servers.
func NewRawMessageEncoder() *RawMessageEncoder {
	return &RawMessageEncoder{}
}

// RawMessageDecoder takes a raw IRC message and parses it into a Message object
type RawMessageDecoder struct{}

func (d *RawMessageDecoder) Decode(s string) (*Message, error) {
	return ParseMessage(s)
}

// NewRawMessageDecoder returns a new RawMessageDecoder object. It is used to
// decode message strings sent by IRC servers into Message objects.
func NewRawMessageDecoder() *RawMessageDecoder {
	return &RawMessageDecoder{}
}

type tenyksChatMessageEncoder struct {
	serverName string
}

func (tme *tenyksChatMessageEncoder) Encode(cmd *PrivmsgCommand) (*servicepb.Message, error) {
	ircMsg := cmd.Message()

	dest := ""
	if len(ircMsg.Params) > 0 {
		channel := ircMsg.Params[0]
		if tme.serverName != "" {
			dest = tme.serverName + "/" + channel
		} else {
			dest = channel
		}
	}

	origin := ""
	if ircMsg.Prefix != nil {
		nick := ircMsg.Prefix.Nick
		if tme.serverName != "" {
			origin = tme.serverName + "/" + nick
		} else {
			origin = nick
		}
	}

	return &servicepb.Message{
		Payload: &servicepb.Message_Chat{
			Chat: &servicepb.Chat{
				DestinationPath: dest,
				OriginPath:      origin,
				Direct:          cmd.IsDirect(),
				Mention:         cmd.IsMention(),
			},
		},
		Content:   ircMsg.Trail,
		CreatedAt: timestamppb.Now(),
	}, nil
}

type tenyksChatMessageDecoder struct{}

func (tmd *tenyksChatMessageDecoder) Decode(msg *servicepb.Message) (*PrivmsgCommand, error) {
	chat := msg.GetChat()
	if chat == nil {
		return nil, errors.New("expected chat message payload")
	}

	_, target := path.Split(chat.DestinationPath)

	return NewPrivmsgCommand(target, msg.Content), nil
}
