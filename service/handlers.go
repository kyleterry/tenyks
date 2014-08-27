package service

import (
	"encoding/json"
	"fmt"
	"time"

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
	serviceMsg.Target = msg.Params[0]
	serviceMsg.Command = msg.Command
	serviceMsg.Mask = msg.Host
	serviceMsg.Direct = irc.IsDirect(msg.Trail, conn.GetCurrentNick())
	serviceMsg.Nick = msg.Nick
	serviceMsg.Host = msg.Host
	serviceMsg.Full_message = msg.RawMsg
	serviceMsg.User = msg.Ident
	serviceMsg.From_channel = irc.IsChannel(msg.Params[0])
	serviceMsg.Connection = conn.Name
	serviceMsg.Meta = &Meta{"Tenyks", TenyksVersion, nil}
	if serviceMsg.Direct {
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
	srv.Online = true
	srv.LastPing = time.Now()
	srv.UUID = meta.SID.UUID
	self.engine.ServiceRg.RegisterService(srv)
}

func (self *Connection) ByeServiceHandler(msg *Message) {
	meta := msg.Meta
	if meta.SID != nil && meta.SID.UUID != nil {
		log.Debug("[service] %s (%s) is hanging up", meta.SID.UUID.String(), meta.Name)
		srv := self.engine.ServiceRg.GetServiceByUUID(meta.Name)
		if srv != nil {
			srv.Online = false
		}
	}
}

type ServiceListMessage struct {
	Services map[string]*Service `json:"services"`
	Command  string              `json:"command"`
	Meta     *Meta               `json:"meta"`
}

func (self *Connection) ListServiceHandler(msg *Message) {
	serviceList := &ServiceListMessage{}
	serviceList.Services = self.engine.ServiceRg.services
	serviceList.Command = "SERVICES"
	serviceList.Meta = &Meta{"Tenyks", TenyksVersion, nil}
	jsonBytes, err := json.Marshal(serviceList)
	if err != nil {
		log.Fatal(err)
	}
	self.Out <- string(jsonBytes[:])
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
