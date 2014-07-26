package service

import (
	"encoding/json"

	"github.com/kyleterry/tenyks/irc"
)

type Message struct {
	Target       string      `json:"target"`
	Command      string      `json:"command"`
	Mask         string      `json:"mask"`
	Direct       bool        `json:"direct"`
	Nick         string      `json:"nick"`
	Host         string      `json:"host"`
	FullMsg      string      `json:"fullmsg"`
	Full_message string      `json:"full_message"` // Legacy for compat with py version
	User         string      `json:"user"`
	FromChannel  bool        `json:"fromchannel"`
	From_channel bool        `json:"from_channel"` // Legacy for compat with py version
	Connection   string      `json:"connection"`
	Payload      string      `json:"payload"`
	Meta         interface{} `json:"meta"`
}

type TenyksMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ServiceMeta struct {
	Name    string `json:"name"`
	Version string `json:"version"`
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
			go handler.Fn(msg)
		}
	}
}

func (self *Connection) dispatch(msg []byte) {
	self.ircify(msg)
}

func dispatch(command string, conn *Connection, msg *Message) {
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
