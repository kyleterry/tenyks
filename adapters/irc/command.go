package irc
// These are used to help map /commands (e.g. /msg foobar sup!?) to their IRC
// friendly commands (e.g. PRIVMSG foobar :sup?!).

import (
	"fmt"
	"errors"
	"strings"
)

type cmdfn func(fullcmd string) string

// This is a map of functions that take a /command, sans /,  and returns
// a string that can be sent to IRC.
var commandHandlers map[string]cmdfn // Really really simple registry. KISS.

func initCommandHandlers() {
	commandHandlers = make(map[string]cmdfn)
	commandHandlers["msg"] = func(fullcmd string) string {
		cmdindex := strings.Index(fullcmd, " ")
		if cmdindex == -1 {
			return ""
		}
		fullcmd = fullcmd[cmdindex+1:]
		targetindex := strings.Index(fullcmd, " ")
		if targetindex == -1 {
			return ""
		}
		target := fullcmd[:targetindex]
		fullcmd = fullcmd[targetindex+1:]
		if len(fullcmd) == 0 {
			return ""
		}
		return fmt.Sprintf("PRIVMSG %s :%s", target, fullcmd)
	}
	commandHandlers["join"] = func(fullcmd string) string {
		cmdindex := strings.Index(fullcmd, " ")
		if cmdindex == -1 {
			return ""
		}
		fullcmd = fullcmd[cmdindex+1:]
		return fmt.Sprintf("JOIN %s", fullcmd)
	}
}

func ConvertSlashCommand(fullcmd string) (string, error) {
	if fullcmd[0] == '/' {
		fullcmd = string(fullcmd[1:]) // get rid of /
		cmdindex := strings.Index(fullcmd, " ")
		cmd := ""
		if cmdindex != -1 {
			cmd = fullcmd[:cmdindex]
		} else {
			cmd = fullcmd
		}
		cmdHandler, ok := commandHandlers[cmd]
		if !ok {
			return "", errors.New(fmt.Sprintf("Command %s not found in map", cmd))
		}
		rawIRCMsg := cmdHandler(fullcmd)
		return rawIRCMsg, nil
	}
	return "", errors.New("Not a slash command")
}
