package irc

import (
	"testing"
)

func TestParseIRCMessage(t *testing.T) {
	t.Parallel()

	cases := []struct {
		description         string
		raw                 string
		shouldParse         bool
		expectedMessageType MessageType
	}{
		{
			description:         "PRIVMSG with prefix",
			raw:                 ":user!~nick@server/mask PRIVMSG #tenyks :yo kyletest2",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "PRIVMSG with v3 tag and prefix",
			raw:                 "@test-tag=test-tag-value;test-tag2 :user!~nick@server/mask PRIVMSG #tenyks :yo kyletest2",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "PRIVMSG with v3 tag and no prefix",
			raw:                 "@test-tag=test-tag-value;test-tag2 PRIVMSG #tenyks :yo kyletest2",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "PRIVMSG with v3 vendor'd tag",
			raw:                 "@+example.com/test-tag=test-tag-value :user!~nick@server/mask PRIVMSG #tenyks :yo kyletest2",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "PRIVMSG with a ':' in the trailing parameter",
			raw:                 ":user!~nick@server/mask PRIVMSG #tenyks :kyle: hey, how's it going?",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "PRIVMSG with a ' :' in the trailing parameter",
			raw:                 ":user!~nick@server/mask PRIVMSG #tenyks :kyle : hey, how's it going?",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "A command with no parameters",
			raw:                 ":user!~nick@server/mask LIST",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "A command that's a response to our JOIN request",
			raw:                 ":user!~nick@server/mask JOIN #tenyks",
			shouldParse:         true,
			expectedMessageType: MessageTypeCommand,
		},
		{
			description:         "A welcome reply",
			raw:                 ":user!~nick@server/mask 001 :Welcome to the IRC server",
			shouldParse:         true,
			expectedMessageType: MessageTypeReply,
		},
		{
			description: "Garbage message 1",
			raw:         "@aklsjdlfkjj",
			shouldParse: false,
		},
		{
			description: "Garbage message 2",
			raw:         ":aklsjdlfkjj",
			shouldParse: false,
		},
	}

	for _, c := range cases {
		t.Run(c.description, func(t *testing.T) {
			msg, err := ParseMessage(c.raw)

			if !c.shouldParse {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if msg.MessageType != c.expectedMessageType {
				t.Errorf("got message type %v, want %v", msg.MessageType, c.expectedMessageType)
			}
		})
	}
}
