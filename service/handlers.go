package service

import (
	"fmt"
	"encoding/json"
	"strings"
	"github.com/kyleterry/tenyks/irc"
)

func PrivmsgHandler(conn *irc.Connection, msg *irc.Message) {
	serviceMsg := Message{}
	serviceMsg.Target = msg.Params[0]
	serviceMsg.Mask = msg.Host
	serviceMsg.Direct = isDirect(msg.Trail, conn)
	serviceMsg.Nick = msg.Nick
	serviceMsg.Host = msg.Host
	serviceMsg.FullMsg = msg.RawMsg
	serviceMsg.Full_message = msg.RawMsg
	serviceMsg.User = msg.Ident
	serviceMsg.FromChannel = isChannel(msg)
	serviceMsg.From_channel = isChannel(msg)
	if serviceMsg.Direct {
		serviceMsg.Payload = stripNickOnDirect(msg.Trail, conn.GetCurrentNick())
	} else {
		serviceMsg.Payload = msg.Trail
	}

	fmt.Printf("%+v\n", serviceMsg)
	jsonStr, err := json.Marshal(serviceMsg)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(jsonStr[:]))
}

func isDirect(msg string, conn *irc.Connection) bool {
	nick := conn.GetCurrentNick()
	possibleDelimeter := string(msg[len(nick)]) // not an off by one. I just happened to need that index.
	if msg[:len(nick)] == nick &&
		(possibleDelimeter == ":" ||
		 possibleDelimeter == "," ||
		 possibleDelimeter == " ") {
		return true
	}
	return false
}

func isChannel(msg *irc.Message) bool {
	if string(msg.Params[0][0]) == "#" {
		return true
	}
	return false
}

func stripNickOnDirect(msg string, nick string) string {
	index := len(nick)
	if string(msg[len(nick) + 1]) == " " {
		index = strings.Index(msg, " ")
	}
	index++
	msg = string(msg[index:])
	return msg
}
