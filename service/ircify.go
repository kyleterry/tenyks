package service

import (
	"encoding/json"

	"code.google.com/p/go-uuid/uuid"
	"github.com/kyleterry/tenyks/irc"
)

type Message struct {
	Target       string `json:"target"`
	Command      string `json:"command"`
	Mask         string `json:"mask"`
	Direct       bool   `json:"direct"`
	Nick         string `json:"nick"`
	Host         string `json:"host"`
	Full_message string `json:"full_message"`
	User         string `json:"user"`
	From_channel bool   `json:"from_channel"`
	Connection   string `json:"connection"`
	Payload      string `json:"payload"`
	Meta         *Meta  `json:"meta"`
}

type Meta struct {
	Name    string    `json:"name"`
	Version string    `json:"version"`
	UUID    uuid.UUID `json:"UUID"`
}

func (self *Connection) ircify(msg []byte) {
	message, err := NewMessageFromBytes(msg)
	if err != nil {
		log.Error("[service] Error parsing message: %s", err)
		return // Just ignore the shit we don't care about
	}
	self.engine.CommandRg.RegistryMu.Lock()
	defer self.engine.CommandRg.RegistryMu.Unlock()
	handlers, ok := self.engine.CommandRg.Handlers[message.Command]
	if ok {
		log.Debug("[service] Dispatching handler `%s`", message.Command)
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(*irc.Handler)
			go handler.Fn(self, message)
		}
	}
}

func (self *Connection) dispatch(msg []byte) {
	self.ircify(msg)
}

func NewMessageFromBytes(msg []byte) (message *Message, err error) {
	message = new(Message)
	jsonerr := json.Unmarshal(msg, &message)
	err = nil
	if jsonerr != nil {
		err = jsonerr
	}
	return
}
