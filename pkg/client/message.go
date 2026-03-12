package client

import (
	"path"
	"strings"

	pb "github.com/kyleterry/tenyks/internal/pb"
)

// Message is a chat message received from tenyks.
type Message struct {
	// Target is where the message was sent (channel name or bot nick for DMs).
	Target string
	// Nick is the sender's nickname.
	Nick string
	// Payload is the text content of the message.
	Payload string
	// Direct is true when the message was sent directly to the bot (DM or mention).
	Direct bool
	// Mention is true when the bot's nick was mentioned in a channel message.
	Mention bool
	// FromChannel is true when the message originated in a channel.
	FromChannel bool
	// Command is the IRC command (always "PRIVMSG" for chat messages).
	Command string
}

func messageFromPB(m *pb.Message) Message {
	chat := m.GetChat()
	target := chat.DestinationPath
	nick := path.Base(chat.OriginPath)
	fromChannel := strings.HasPrefix(path.Base(target), "#")

	return Message{
		Target:      target,
		Nick:        nick,
		Payload:     m.Content,
		Direct:      chat.Direct,
		Mention:     chat.Mention,
		FromChannel: fromChannel,
		Command:     "PRIVMSG",
	}
}
