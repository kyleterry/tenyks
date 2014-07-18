package service

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Target       string `json:"target"`
	Mask         string `json:"mask"`
	Direct       bool   `json:"direct"`
	Nick         string `json:"nick"`
	Host         string `json:"host"`
	FullMsg      string `json:"fullmsg"`
	Full_message string `json:"full_message"` // Legacy for compat with py version
	User         string `json:"user"`
	FromChannel  bool   `json:"fromchannel"`
	From_channel bool   `json:"from_channel"` // Legacy for compat with py version
	Payload      string `json:"payload"`
}

func ircify(msg []byte) {
	message, err := NewMessage(msg)
	if err != nil {
		log.Error("[service] Error parsing message: %s", err)
		return // Just ignore the shit we don't care about
	}
	fmt.Printf("%+v\n", message)
}

func NewMessage(msg []byte) (message *Message, err error) {
	message = new(Message)
	jsonerr := json.Unmarshal(msg, &message)
	err = nil
	if jsonerr != nil {
		err = jsonerr
	}
	return
}
