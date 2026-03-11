package irc

import (
	"errors"
	"fmt"
	"strings"
)

// Command is a wrapper for Message that can be recieved and sent. It
// implements MessageObject.
type Command interface {
	MessageObject
	Encode() (string, error)
	Validate() error
}

// DefaultCommandFactory is a factory that maps CommandTypes to functions that return new
// Command implementations. The factory functions take message and instantiate
// Commands with that message.
var DefaultCommandFactory = map[CommandType]func(*Message) Command{
	CommandTypePass: func(msg *Message) Command {
		return &PassCommand{m: msg}
	},
	CommandTypeUser: func(msg *Message) Command {
		return &UserCommand{m: msg}
	},
	CommandTypeNick: func(msg *Message) Command {
		return &NickCommand{m: msg}
	},
	CommandTypeJoin: func(msg *Message) Command {
		return &JoinCommand{m: msg}
	},
	CommandTypePrivmsg: func(msg *Message) Command {
		return &PrivmsgCommand{m: msg}
	},
	CommandTypePing: func(msg *Message) Command {
		return &PingCommand{m: msg}
	},
	CommandTypePong: func(msg *Message) Command {
		return &PongCommand{m: msg}
	},
}

type PassCommand struct {
	m *Message
}

func (p PassCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(p.m)
}

func (p PassCommand) Message() *Message {
	return p.m
}

func (p PassCommand) Validate() error {
	if len(p.m.Params) != 1 {
		return errors.New("PASS command: wrong number of parameters")
	}

	return nil
}

func NewPassCommand(password string) *PassCommand {
	return &PassCommand{
		m: &Message{
			Command:     "PASS",
			MessageType: MessageTypeCommand,
			Params:      []string{password},
		},
	}
}

type UserCommand struct {
	m *Message
}

func (u UserCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(u.m)
}

func (u UserCommand) Message() *Message {
	return u.m
}

func (u UserCommand) Validate() error {
	if len(u.m.Params) != 3 {
		return errors.New("USER command error: wrong number of parameters")
	}

	if u.m.Trail == "" {
		return errors.New("USER command error: real name parameter is required")
	}

	return nil
}

func NewUserCommand(user string, mode int, realName string) *UserCommand {
	return &UserCommand{
		m: &Message{
			Command:     "USER",
			MessageType: MessageTypeCommand,
			Params:      []string{user, fmt.Sprintf("%d", mode), "*"},
			Trail:       realName,
		},
	}
}

type NickCommand struct {
	m *Message
}

func (n NickCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(n.m)
}

func (n NickCommand) Message() *Message {
	return n.m
}

func (n NickCommand) Validate() error {
	if len(n.m.Params) != 1 {
		return errors.New("NICK command: nick string parameter is required")
	}

	return nil
}

func NewNickCommand(nick string) *NickCommand {
	return &NickCommand{
		m: &Message{
			Command:     "NICK",
			MessageType: MessageTypeCommand,
			Params:      []string{nick},
		},
	}
}

type JoinCommand struct {
	m *Message
}

func (j JoinCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(j.m)
}

func (j JoinCommand) Message() *Message {
	return j.m
}

func (j JoinCommand) Validate() error {
	if len(j.m.Params) < 1 {
		return errors.New("JOIN command: channels parameter is required")
	}

	return nil
}

func NewJoinCommand(channels ...string) *JoinCommand {
	// TODO validate channel name
	return &JoinCommand{
		m: &Message{
			Command:     "JOIN",
			MessageType: MessageTypeCommand,
			Params:      []string{strings.Join(channels, ",")},
		},
	}
}

type PrivmsgCommand struct {
	m             *Message
	isDirectFunc  func(*Message) bool
	isMentionFunc func(*Message) bool
}

func (p PrivmsgCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(p.m)
}

func (p PrivmsgCommand) Message() *Message {
	return p.m
}

func (p PrivmsgCommand) Validate() error {
	if len(p.m.Params) != 1 {
		return errors.New("PRIVMSG command: wrong number of parameters")
	}

	if p.m.Trail == "" {
		return errors.New("PRIVMSG command: a message is required")
	}

	return nil
}

func (p PrivmsgCommand) IsDirect() bool {
	if p.isDirectFunc == nil {
		return false
	}

	return p.isDirectFunc(p.m)
}

func (p PrivmsgCommand) IsMention() bool {
	if p.isMentionFunc == nil {
		return false
	}

	return p.isMentionFunc(p.m)
}

func NewPrivmsgCommand(target string, msg string) *PrivmsgCommand {
	return &PrivmsgCommand{
		m: &Message{
			Command:     "PRIVMSG",
			MessageType: MessageTypeCommand,
			Params:      []string{target},
			Trail:       msg,
		},
	}
}

func mentionAndDirectPrivmsgCommand(c *Connection, m *Message) Command {
	return &PrivmsgCommand{
		m: m,
		isDirectFunc: func(m *Message) bool {
			if c.Status.Connected == false {
				return false
			}

			return m.Params[0] == c.Status.CurrentNick
		},
		isMentionFunc: func(m *Message) bool {
			if c.Status.Connected == false {
				return false
			}

			parts := strings.Split(m.Trail, ": ")
			if len(parts) < 2 {
				return false
			}

			return parts[0] == c.Status.CurrentNick
		},
	}
}

type PingCommand struct {
	m *Message
}

func (p PingCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(p.m)
}

func (p PingCommand) Message() *Message {
	return p.m
}

func (p PingCommand) Validate() error {
	return nil
}

type PongCommand struct {
	m *Message
}

func (p PongCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(p.m)
}

func (p PongCommand) Message() *Message {
	return p.m
}

func (p PongCommand) Validate() error {
	return nil
}

func NewPongCommand(server string) *PongCommand {
	return &PongCommand{
		m: &Message{
			Command:     "PONG",
			MessageType: MessageTypeCommand,
			Trail:       server,
		},
	}
}

type UnknownCommand struct {
	m *Message
}

func (c UnknownCommand) Encode() (string, error) {
	return NewRawMessageEncoder().Encode(c.m)
}

func (c UnknownCommand) Message() *Message {
	return c.m
}

func (c UnknownCommand) Validate() error {
	return nil
}
