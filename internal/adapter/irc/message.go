package irc

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

// MessageType is the type of message received from the IRC server.
type MessageType int

const (
	// MessageTypeCommand is a message that instructs the server or client to
	// do something, such as join a channel or issue a pong response to a ping.
	MessageTypeCommand MessageType = iota
	// MessageTypeReply is a message that serves as a response to a command,
	// such as a list of names that are members of a channel or an error
	// related to an attempt to join a channel.
	MessageTypeReply
)

// MessageObject is an interface all message types must implement.
type MessageObject interface {
	Message() *Message
}

// ParseMessage takes a string from an IRC server and uses the following BNF from the
// RFC to parse it into a message.Message. Keep in mind that this BNF is a combination
// of RFC2812 and IRCv3 augmentation.
//
// <message>       ::= ['@' <tags> <SPACE>] [':' <prefix> <SPACE> ] <command> <params> <crlf>
// <tags>          ::= <tag> [';' <tag>]*
// <tag>           ::= <key> ['=' <escaped_value>]
// <key>           ::= [ <client_prefix> ] [ <vendor> '/' ] <key_name>
// <client_prefix> ::= '+'
// <key_name>      ::= <sequence of letters, digits, hyphens ('-')>
// <escaped_value> ::= <sequence of any characters except NUL, CR, LF, semicolon (`;`) and SPACE>
// <vendor>        ::= <host>
// <prefix>        ::= <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
// <command>       ::= <letter> { <letter> } | <number> <number> <number>
// <SPACE>         ::= ' ' { ' ' }
// <params>        ::= <SPACE> [ ':' <trailing> | <middle> <params> ]
// <middle>        ::= <Any *non-empty* sequence of octets not including SPACE
//
//	or NUL or CR or LF, the first of which may not be ':'>
//
// <trailing>      ::= <Any, possibly *empty*, sequence of octets not including
//
//	NUL or CR or LF>
//
// <crlf>          ::= CR LF
//
// RFC2812: https://tools.ietf.org/html/rfc2812
func ParseMessage(raw string) (*Message, error) {
	raw = strings.TrimSuffix(raw, "\r\n")

	msg := Message{
		CreatedAt: time.Now(),
		RawMsg:    raw,
	}

	// If the message starts with @, then we will be parsing an IRC v3 message that
	// has tags in it
	if raw[0] == '@' {
		if index := strings.Index(raw, " "); index != -1 {
			if msg.Tags == nil {
				msg.Tags = &TagsSection{}
			}

			rawTags := raw[1:index]
			msg.Tags.RawTags = rawTags
			raw = raw[index+1:]

			for _, tag := range strings.Split(rawTags, ";") {
				var key, value, vendor string

				parts := strings.Split(tag, "=")
				if len(parts) > 2 {
					return nil, errors.New("failed to parse message: invalid tag format")
				}

				rawKey := parts[0]

				if len(parts) > 1 {
					value = parts[1]
				}

				if rawKey[0] == '+' {
					rawKey = rawKey[1:]

					if index := strings.Index(rawKey, "/"); index != -1 {
						vendor = rawKey[:index]
						rawKey = rawKey[index+1:]
					}
				}

				key = rawKey

				msg.Tags.Tags = append(msg.Tags.Tags, &Tag{
					Key:    key,
					Value:  value,
					Vendor: vendor,
				})
			}
		} else {
			return nil, errors.New("failed to parse message: invalid message format")
		}
	}

	// If the next section starts with :, then we will be parsing a prefix section
	if raw[0] == ':' {
		var rawPrefix string

		if msg.Prefix == nil {
			msg.Prefix = &PrefixSection{}
		}

		// We have what looks to be a prefixed message from IRC, lets start trying to
		// parse the prefix section
		if index := strings.Index(raw, " "); index != -1 { // fetch up to " "
			rawPrefix = raw[1:index]
			msg.Prefix.RawPrefix = rawPrefix // could be server or user string
			raw = raw[index+1:]
		} else {
			// If we start with a : and there's no <space> anywhere, then we have an
			// invalid message
			return nil, errors.New("failed to parse message: invalid prefix")
		}

		// Now we can search the prefix for ! and @ to see if we have nick and user
		// data. If we don't then we just ignore it and the prefix only contains the server
		// name
		nickIndex := strings.Index(rawPrefix, "!")
		userIndex := strings.Index(rawPrefix, "@")
		if nickIndex != -1 && userIndex != -1 {
			msg.Prefix.Nick = rawPrefix[:nickIndex]
			msg.Prefix.Ident = rawPrefix[nickIndex+1 : userIndex]
			msg.Prefix.Host = rawPrefix[userIndex+1:]
		}
	}

	// Split at the first instance of " :" to parse out the trailing parameter
	// (a param who's text can contain a space).
	parts := strings.SplitN(raw, " :", 2)
	params := strings.Fields(parts[0])

	// If there are 2 elements, then we have the params list as a string and the
	// trailing parameter as a string, so we need to store the trailing param as well.
	if len(parts) == 2 {
		msg.Trail = parts[1]
	}

	msg.Command = params[0]

	// Detect the message type according to the RFC. Default to the Command
	// message type, but if the command is 3 characters and can be converted to
	// an integer, it means we have a reply code and we should set the message
	// type to Reply.
	// https://tools.ietf.org/html/rfc2812#section-5
	msg.MessageType = MessageTypeCommand
	if len(msg.Command) == 3 {
		if _, err := strconv.Atoi(msg.Command); err == nil {
			msg.MessageType = MessageTypeReply
		}
	}

	// Finally pop the command off the params list and then store the remaining
	// params if there are any.
	if len(params) > 1 {
		msg.Params = params[1:]
	}

	msg.Parsed = true

	return &msg, nil
}

// Message is a type that holds all the parsed information from a message string
type Message struct {
	// Tags holds IRC V3 tags if there were any in the message
	Tags *TagsSection
	// Prefix holds the message prefix (from server only). Tenyks should never
	// send one of these to a server and it will be ignored by the MessageEncoder.
	Prefix *PrefixSection
	// Command is the IRC command (also includes reply codes and error codes)
	Command string
	// MessageType is the type of message received. This field can be inspected
	// on a MessageObject to skip a type switch step when asserting to a
	// concrete command or reply object.
	MessageType MessageType
	// Params are the parameters for the command
	Params []string
	// Trail is a parameter that starts with : used as a syntactic trick to allow
	// a parameter to have a <SPACE> character.
	Trail string
	// CreatedAt is when we recieved the message and parsed it.
	CreatedAt time.Time
	// RawMessage is the full unaltered message.
	RawMsg string
	// Parsed is a boolean flag for detecting if we used ParseMessage
	Parsed bool
}

type TagsSection struct {
	// Tags are a list of Tag objects
	Tags []*Tag
	// RawTags is the unaltered tag string recieved
	RawTags string
}

type Tag struct {
	// Key is the required tag key
	Key string
	// Value is an optional tag value
	Value string
	// Vendor is an optional vendor identifier
	Vendor string
}

type PrefixSection struct {
	// Nick is the user nick or server name that represents the source of a parsed
	// message
	Nick string
	// Ident is the IDENT the irc server has for this message source
	Ident string
	// Host is the host where the message originated from
	Host string
	// RawPrefix is the unaltered prefix string recieved
	RawPrefix string
}
