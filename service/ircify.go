package service

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Target       string
	Mask         string
	Direct       bool
	Nick         string
	Host         string
	FullMessage  string
	Full_message string // Legacy for compat with py version
	User         string
	FromChannel  bool
	From_channel bool // Legacy for compat with py version
	Payload      string
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
