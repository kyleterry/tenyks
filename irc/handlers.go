package irc

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type HandlerRegistry struct {
	Handlers   map[string]*list.List
	RegistryMu *sync.Mutex
}

func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		Handlers: make(map[string]*list.List),
		RegistryMu: &sync.Mutex{},
	}
}

func (h *HandlerRegistry) AddHandler(name string, handler *Handler) {
	h.RegistryMu.Lock()
	defer h.RegistryMu.Unlock()
	if _, ok := h.Handlers[name]; !ok {
		h.Handlers[name] = list.New()
	}
	h.Handlers[name].PushBack(handler)
}

type Handler struct {
	Fn func(...interface{})
}

func NewHandler(fn func(...interface{})) *Handler {
	return &Handler{fn}
}

type ircfn func(*Connection, *Message)

func (conn *Connection) AddHandler(name string, fn ircfn) {
	handler := NewHandler(func(p ...interface{}) {
		fn(p[0].(*Connection), p[1].(*Message))
	})
	conn.Registry.AddHandler(name, handler)
}

func (conn *Connection) addBaseHandlers() {
	conn.AddHandler("bootstrap", (*Connection).BootstrapHandler)
	conn.AddHandler("send_ping", (*Connection).SendPing)
	conn.AddHandler("001", (*Connection).ConnectedHandler)
	conn.AddHandler("433", (*Connection).NickInUseHandler)
	conn.AddHandler("PING", (*Connection).PingHandler)
	conn.AddHandler("PONG", (*Connection).PongHandler)
	conn.AddHandler("CTCP", (*Connection).CTCPHandler)
}

func (conn *Connection) PingHandler(msg *Message) {
	log.Debug("[%s] Responding to PING", conn.Name)
	conn.Out <- fmt.Sprintf("PONG %s", msg.Trail)
}

func (conn *Connection) PongHandler(msg *Message) {
	conn.LastPong = time.Now()
	conn.PongIn <- true
}

func (conn *Connection) SendPing(msg *Message) {
	log.Debug("[%s] Sending PING to server %s", conn.Name, conn.currentServer)
	conn.Out <- fmt.Sprintf("PING %s", conn.currentServer)
}

func (conn *Connection) BootstrapHandler(msg *Message) {
	log.Info("[%s] Bootstrapping connection", conn.Name)
	conn.Out <- fmt.Sprintf(
		"USER %s %s %s :%s",
		conn.Config.Nicks[0],
		conn.Config.Host,
		conn.Config.Ident,
		conn.Config.Realname)
	conn.Out <- fmt.Sprintf(
		"NICK %s", conn.Config.Nicks[conn.nickIndex])
	conn.currentNick = conn.Config.Nicks[conn.nickIndex]
	conn.ConnectWait <- true
	close(conn.ConnectWait)
}

func (conn *Connection) NickInUseHandler(msg *Message) {
	log.Info("[%s] Nick `%s` is in use. Next...", conn.Name, conn.currentNick)
	conn.nickIndex++
	if len(conn.Config.Nicks) >= conn.nickIndex+1 {
		conn.Out <- fmt.Sprintf(
			"NICK %s", conn.Config.Nicks[conn.nickIndex])
		conn.currentNick = conn.Config.Nicks[conn.nickIndex]
	} else {
		log.Fatal("All nicks in use.")
	}
}

func (conn *Connection) ConnectedHandler(msg *Message) {
	log.Info("[%s] Sending user commands", conn.Name)
	initCommandHandlers()
	for _, commandHook := range conn.Config.Commands {
		ircsafe, err := ConvertSlashCommand(commandHook)
		if err != nil { // If there's an error, just try to send commandHook
			ircsafe = commandHook
		}
		conn.Out <- ircsafe
	}
	log.Info("[%s] Joining Channels", conn.Name)
	// TODO: generalize joins
	for _, channel := range conn.Config.Channels {
		conn.Out <- fmt.Sprintf("JOIN %s", channel)
		log.Debug("[%s] Joined %s", conn.Name, channel)
	}
	conn.currentServer = msg.Prefix
	go conn.watchdog()
}

func (conn *Connection) CTCPHandler(msg *Message) {

}
