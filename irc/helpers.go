package irc

import (
	"strings"
)

func IsDirect(msg string, conn *Connection) bool {
	nick := conn.GetCurrentNick()
	if len(msg) > len(nick) {
		possibleDelimeter := string(msg[len(nick)]) // not an off by one. I just happened to need that index.
		if msg[:len(nick)] == nick &&
			(possibleDelimeter == ":" ||
				possibleDelimeter == "," ||
				possibleDelimeter == " ") {
			return true
		}
	}
	return false
}

func IsChannel(msg *Message) bool {
	if string(msg.Params[0][0]) == "#" {
		return true
	}
	return false
}

func StripNickOnDirect(msg string, nick string) string {
	index := len(nick)
	if string(msg[len(nick)+1]) == " " {
		index = strings.Index(msg, " ")
	}
	index++
	msg = string(msg[index:])
	return msg
}
