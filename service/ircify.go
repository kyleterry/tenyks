package service

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/Xe/uuid"
	"github.com/kyleterry/tenyks/irc"
)

type Message struct {
	Target       string `json:"target"`
	Command      string `json:"command"`
	Mask         string `json:"mask"`
	Direct       bool   `json:"direct"`
	Nick         string `json:"nick"`
	Host         string `json:"host"`
	Full_message string `json:"full_message"` // Legacy for compat with py version
	User         string `json:"user"`
	From_channel bool   `json:"from_channel"` // Legacy for compat with py version
	Connection   string `json:"connection"`
	Payload      string `json:"payload"`
	Meta         *Meta  `json:"meta"`
}

type ServiceID struct {
	UUID uuid.UUID
}

type Meta struct {
	Name        string     `json:"name"`
	Version     string     `json:"version"`
	SID         *ServiceID `json:"UUID"`
	Description string     `json:"description"`
}

func (self *ServiceID) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), "\"")
	self.UUID = uuid.Parse(s)
	if self.UUID == nil {
		return errors.New("Could not parse UUID")
	}
	return nil
}

func (self *Connection) ircify(msg string) {
	message, err := NewMessageFromString(msg)
	if err != nil {
		Logger.Error("error parsing message", "error", err)
		return // Just ignore the shit we don't care about
	}
	self.engine.CommandRg.RegistryMu.Lock()
	defer self.engine.CommandRg.RegistryMu.Unlock()
	handlers, ok := self.engine.CommandRg.Handlers[message.Command]
	if ok {
		Logger.Debug("dispatching handler", "command", message.Command)
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(*irc.Handler)
			go handler.Fn(self, message)
		}
	}
}

func (self *Connection) dispatch(msg string) {
	self.ircify(msg)
}

func NewMessageFromString(msg string) (message *Message, err error) {
	message = new(Message)
	jsonerr := json.Unmarshal([]byte(msg), &message)
	err = nil
	if jsonerr != nil {
		err = jsonerr
	}
	return
}
