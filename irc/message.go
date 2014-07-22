package irc

import (
	"regexp"
	"strings"
	"time"
)

// Extend Regexp for map support
type ircRegexp struct {
	*regexp.Regexp
}

// Maps all captured names and their values
func (r *ircRegexp) FindStringSubmatchMap(s string) map[string]string {
	matches := make(map[string]string)

	match := r.FindStringSubmatch(s)
	if match == nil {
		return matches
	}

	for i, name := range r.SubexpNames() {
		if i == 0 || name == "" {
			continue
		}
		matches[name] = match[i]
	}
	return matches
}

type Message struct {
	Prefix  string
	Nick    string
	Ident   string
	Host    string
	Command string
	Trail   string
	Params  []string
	SentAt  time.Time
	RawMsg  string
	Conn    *Connection
}

func (m *Message) String() string {
	return ""
}

// Parses a message string from an IRC server
func ParseMessage(rawMsg string) *Message {
	/*
		<message>  ::= [':' <prefix> <SPACE> ] <command> <params> <crlf>
		<prefix>   ::= <servername> | <nick> [ '!' <user> ] [ '@' <host> ]
		<command>  ::= <letter> { <letter> } | <number> <number> <number>
		<SPACE>    ::= ' ' { ' ' }
		<params>   ::= <SPACE> [ ':' <trailing> | <middle> <params> ]

		<middle>   ::= <Any *non-empty* sequence of octets not including SPACE
						or NUL or CR or LF, the first of which may not be ':'>
		<trailing> ::= <Any, possibly *empty*, sequence of octets not including
						NUL or CR or LF>
		<crlf>     ::= CR LF
	*/
	rawMsg = strings.TrimSuffix(rawMsg, "\r\n")
	msg := Message{
		SentAt: time.Now(),
		RawMsg: rawMsg,
	}
	if rawMsg[0] == ':' { // Valid IRC message from server
		if index := strings.Index(rawMsg, " "); index != -1 { // fetch up to " "
			msg.Prefix = rawMsg[1:index] // could be server or user string
			rawMsg = rawMsg[index+1:]
		} else {
			return nil //lulwut?
		}
		nickIndex := strings.Index(msg.Prefix, "!")
		userIndex := strings.Index(msg.Prefix, "@")
		if nickIndex != -1 && userIndex != -1 {
			msg.Nick = msg.Prefix[:nickIndex]
			msg.Ident = msg.Prefix[nickIndex+1 : userIndex]
			msg.Host = msg.Prefix[userIndex+1:]
		}
		// Done with prefix
	} else if rawMsg[0] == ' ' {
		return nil
	}

	tmpCommand := strings.SplitN(rawMsg, " :", 2)
	if len(tmpCommand) > 1 { // There seems to be a command, args and a trail
		msg.Params = strings.Fields(tmpCommand[0])
		msg.Trail = tmpCommand[1]
	} else { // No trail and only a command
		msg.Params = strings.Fields(tmpCommand[0])
	}
	msg.Command = msg.Params[0] // "pop" off the command
	if len(msg.Params) > 1 {
		msg.Params = msg.Params[1:] // and store the remaining params
	}
	return &msg
}
