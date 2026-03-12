package client

// PrivmsgMiddleware wraps a Handler so it only fires for PRIVMSG messages.
func PrivmsgMiddleware(handler Handler) Handler {
	return HandlerFunc(func(r Result, msg Message, com *Communication) {
		if msg.Command == "PRIVMSG" {
			handler.HandleMatch(r, msg, com)
		}
	})
}
