package service

import (
	"fmt"
	"encoding/json"
)

type Message struct {
	Target string
	Mask string
	Direct bool
	Nick string
	Host string
	FullMessage string
	Full_message string
	User string
	FromChannel bool
	From_channel bool
	Payload string
}

func ircify(msg []byte) {
	message, err := NewMessage(msg)
	if err != nil {
		log.Error("[service] Error parsing message: %s", err)
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
