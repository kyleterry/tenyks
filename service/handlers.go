package service

import (
	"encoding/json"
	"fmt"
	"time"
	"strings"

	"github.com/kyleterry/tenyks/irc"
	. "github.com/kyleterry/tenyks/version"
)

type servicefn func(*Connection, *Message)

func (self *ServiceEngine) AddHandler(name string, fn servicefn) {
	handler := irc.NewHandler(func(p ...interface{}) {
		fn(p[0].(*Connection), p[1].(*Message))
	})
	self.CommandRg.AddHandler(name, handler)
}

func (self *ServiceEngine) addBaseHandlers() {
	self.AddHandler("PRIVMSG", (*Connection).PrivmsgServiceHandler)
	self.AddHandler("REGISTER", (*Connection).RegisterServiceHandler)
	self.AddHandler("BYE", (*Connection).ByeServiceHandler)
	self.AddHandler("PONG", (*Connection).PongServiceHandler)
}

func (self *Connection) PrivmsgIrcHandler(conn *irc.Connection, msg *irc.Message) {
	serviceMsg := Message{}
	if !irc.IsChannel(msg.Params[0]) {
		serviceMsg.Target = msg.Nick
	} else {
		serviceMsg.Target = msg.Params[0]
	}
	serviceMsg.Command = msg.Command
	serviceMsg.Mask = msg.Host
	if irc.IsDirect(msg.Trail, conn.GetCurrentNick()) || !irc.IsChannel(msg.Params[0]) {
		serviceMsg.Direct = true
	} else {
		serviceMsg.Direct = false
	}
	serviceMsg.Nick = msg.Nick
	serviceMsg.Host = msg.Host
	serviceMsg.Full_message = msg.RawMsg
	serviceMsg.User = msg.Ident
	serviceMsg.From_channel = irc.IsChannel(msg.Params[0])
	serviceMsg.Connection = conn.Name
	serviceMsg.Meta = &Meta{"Tenyks", TenyksVersion, nil, ""}
	if serviceMsg.Direct && serviceMsg.From_channel {
		serviceMsg.Payload = irc.StripNickOnDirect(msg.Trail, conn.GetCurrentNick())
	} else {
		serviceMsg.Payload = msg.Trail
	}

	jsonBytes, err := json.Marshal(serviceMsg)
	if err != nil {
		log.Fatal(err)
	}
	self.Out <- string(jsonBytes[:])
}

func (self *Connection) ListServicesIrcHandler(conn *irc.Connection, msg *irc.Message) {
	if irc.IsDirect(msg.Trail, conn.GetCurrentNick()) {
		if strings.Contains(msg.RawMsg, "!services") {
			log.Debug("[service] List services triggered")
			if len(self.engine.ServiceRg.services) > 0 {
				for _, service := range self.engine.ServiceRg.services {
					outMessage := fmt.Sprintf("%s", service)
					conn.Out <- msg.GetDMString(outMessage)
				}
			} else {
				conn.Out <- msg.GetDMString("No services registered")
			}
		}
	}
}

func (self *Connection) HelpIrcHandler(conn *irc.Connection, msg *irc.Message) {
	if irc.IsDirect(msg.Trail, conn.GetCurrentNick()) {
		trail := irc.StripNickOnDirect(msg.Trail, conn.GetCurrentNick())
		if strings.HasPrefix(trail, "!help") {
			trail_pieces := strings.Fields(trail)
			if len(trail_pieces) > 1 {
				if self.engine.ServiceRg.IsService(trail_pieces[1]) {
					service := self.engine.ServiceRg.GetServiceByName(trail_pieces[1])
					if service == nil {
						conn.Out <- msg.GetDMString(
							fmt.Sprintf("No such service `%s`", trail_pieces[1]))
						return
					}
					serviceMsg := &Message{
						Target: msg.Nick,
						Nick: msg.Nick,
						Direct: true,
						From_channel: false,
						Command: "PRIVMSG",
						Connection: conn.Name,
						Payload: fmt.Sprintf("!help %s", service.UUID.String()),
					}
					jsonBytes, err := json.Marshal(serviceMsg)
					if err != nil {
						log.Error("Cannot marshal help message")
						return
					}
					self.Out <- string(jsonBytes[:])
				} else {
					conn.Out <- msg.GetDMString(
						fmt.Sprintf("No such service `%s`", trail[1]))
				}
			} else {
				conn.Out <- msg.GetDMString(
					fmt.Sprintf("%s: !help - This help message", conn.GetCurrentNick()))
				conn.Out <- msg.GetDMString(
					fmt.Sprintf("%s: !services - List services", conn.GetCurrentNick()))
				conn.Out <- msg.GetDMString(
					fmt.Sprintf("%s: !help <servicename> - Get help for a service", conn.GetCurrentNick()))
			}
		}
	}
}

func (self *Connection) PrivmsgServiceHandler(msg *Message) {
	conn := self.getIrcConnByName(msg.Connection)
	if conn != nil {
		msgStr := fmt.Sprintf("%s %s :%s", msg.Command, msg.Target, msg.Payload)
		conn.Out <- msgStr
	} else {
		log.Debug("[service] No such connection `%s`. Ignoring.",
			msg.Connection)
	}
}

func (self *Connection) RegisterServiceHandler(msg *Message) {
	meta := msg.Meta
	if meta.SID == nil || meta.SID.UUID == nil {
		log.Error("[service] ERROR: UUID required to register with Tenyks")
		return
	}
	log.Debug("[service] %s (%s) wants to register", meta.SID.UUID.String(), meta.Name)
	srv := &Service{}
	srv.Name = meta.Name
	srv.Version = meta.Version
	srv.Description = meta.Description
	srv.Online = true
	srv.LastPing = time.Now()
	srv.UUID = meta.SID.UUID
	self.engine.ServiceRg.RegisterService(srv)
}

func (self *Connection) ByeServiceHandler(msg *Message) {
	meta := msg.Meta
	if meta.SID != nil && meta.SID.UUID != nil {
		log.Debug("[service] %s (%s) is hanging up", meta.SID.UUID.String(), meta.Name)
		srv := self.engine.ServiceRg.GetServiceByUUID(meta.SID.UUID.String())
		if srv != nil {
			log.Debug("[service] Settings state to `offline` for `%s`", srv.Name)
			srv.Online = false
		}
	}
}

const (
	ServiceOnline = true
	ServiceOffline = false
)

func (self *Connection) PingServices() {
	log.Debug("[service] Starting pinger")
	for {
		<-time.After(time.Second * 120)
		log.Debug("[service] PINGing services")
		msg := &Message{
			Command: "PING",
			Payload: "!tenyks",
		}
		jsonBytes, err := json.Marshal(msg)
		if err != nil {
			log.Error("Cannot marshal PING message")
			continue
		}
		self.Out <- string(jsonBytes[:])

		services := self.engine.ServiceRg.services
		for _, service := range services {
			if service.Online {
				service.LastPing = time.Now()
			}
		}
	}
}

func (self *Connection) PongServiceHandler(msg *Message) {
	meta := msg.Meta
	if meta.SID != nil && meta.SID.UUID != nil {
		self.engine.UpdateService(meta.SID.UUID.String(), ServiceOnline)
	}
}
