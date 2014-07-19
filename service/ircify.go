package service

import (
	"encoding/json"
	"fmt"
	"github.com/kyleterry/tenyks/irc"
)

type Message struct {
	Target       string `json:"target"`
	Command      string `json:"command"`
	Mask         string `json:"mask"`
	Direct       bool   `json:"direct"`
	Nick         string `json:"nick"`
	Host         string `json:"host"`
	FullMsg      string `json:"fullmsg"`
	Full_message string `json:"full_message"` // Legacy for compat with py version
	User         string `json:"user"`
	FromChannel  bool   `json:"fromchannel"`
	From_channel bool   `json:"from_channel"` // Legacy for compat with py version
	Connection   string `json:"connection"`
	Payload      string `json:"payload"`
}

func (self *Connection) ircify(msg []byte) {
	message, err := NewMessageFromBytes(msg)
	if err != nil {
		log.Error("[service] Error parsing message: %s", err)
		return // Just ignore the shit we don't care about
	}
	if message.Command == "PRIVMSG" {
		conn := self.getIrcConnByName(message.Connection)
		if conn != nil {
			msgStr := fmt.Sprintf("%s %s :%s", message.Command, message.Target, message.Payload)
			conn.Out <- msgStr
		} else {
			log.Debug("[service] No such connection `%s`. Ignoring.",
				message.Connection)
		}
	}
}

func (self *Connection) getIrcConnByName(name string) *irc.Connection {
	conn := self.ircconns[name]
	if conn == nil {
		log.Error("[service] Connection `%s` doesn't exist", name)
	}
	return conn
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
