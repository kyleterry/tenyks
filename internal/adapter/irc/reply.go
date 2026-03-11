package irc

import (
	"fmt"
	"strings"
)

type Reply interface {
	MessageObject
	Validate() error
}

var DefaultReplyFactory = map[ReplyType]func(*Message) Reply{
	ReplyTypeWelcome: func(msg *Message) Reply {
		return &WelcomeReply{m: msg}
	},
	ReplyTypeNames: func(msg *Message) Reply {
		return &NamesReply{m: msg}
	},
	ReplyTypeEndOfNames: func(msg *Message) Reply {
		return &EndOfNamesReply{m: msg}
	},
	ReplyTypeErrNickInUse: func(msg *Message) Reply {
		return &ErrNickInUseReply{m: msg}
	},
}

type WelcomeReply struct {
	m *Message
}

func (r WelcomeReply) Message() *Message {
	return r.m
}

func (r WelcomeReply) Validate() error {
	return nil
}

type NamesReply struct {
	m *Message
}

func (r NamesReply) Message() *Message {
	return r.m
}

func (r NamesReply) Validate() error {
	l := len(r.m.Params)
	if l < 2 || l > 3 {
		return fmt.Errorf("%w: expected 2 or 3, but got %d", ParameterCountValidationError, l)
	}

	return nil
}

func (r NamesReply) Names() []string {
	return strings.Split(r.m.Trail, " ")
}

func (r NamesReply) Channel() string {
	// If we don't have params, this will blow up. That means the caller didn't
	// validate the reply first and it's their fault :^).
	return r.m.Params[len(r.m.Params)-1]
}

type EndOfNamesReply struct {
	m *Message
}

func (r EndOfNamesReply) Message() *Message {
	return r.m
}

func (r EndOfNamesReply) Validate() error {
	if len(r.m.Params) == 2 {
		return fmt.Errorf("%w: expected 2, but got %d", ParameterCountValidationError, len(r.m.Params))
	}

	return nil
}

func (r EndOfNamesReply) Channel() string {
	return r.m.Params[len(r.m.Params)-1]
}

type ErrNickInUseReply struct {
	m *Message
}

func (r ErrNickInUseReply) Message() *Message {
	return r.m
}

func (r ErrNickInUseReply) Validate() error {
	return nil
}
