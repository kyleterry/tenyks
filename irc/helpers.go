package irc

import (
	"strings"
)

func IsDirect(msg string, currentNick string) bool {
	nick := currentNick
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

func IsChannel(target string) bool {
	if string(target[0]) == "#" {
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
