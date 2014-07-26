package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyleterry/tenyks/irc"
	. "github.com/kyleterry/tenyks/version"
)

type servicefn func(*Message)

func (self *ServiceEngine) AddHandler(name string, fn servicefn) {
	handler := irc.NewHandler(func(p ...interface{}) {
		fn(p[0].(*Message))
	})
	self.CommandRg.AddHandler(name, handler)
}

func (self *Connection) PrivmsgIrcHandler(conn *irc.Connection, msg *irc.Message) {
	serviceMsg := Message{}
	serviceMsg.Target = msg.Params[0]
	serviceMsg.Command = msg.Command
	serviceMsg.Mask = msg.Host
	serviceMsg.Direct = irc.IsDirect(msg.Trail, conn)
	serviceMsg.Nick = msg.Nick
	serviceMsg.Host = msg.Host
	serviceMsg.FullMsg = msg.RawMsg
	serviceMsg.Full_message = msg.RawMsg
	serviceMsg.User = msg.Ident
	serviceMsg.FromChannel = irc.IsChannel(msg)
	serviceMsg.From_channel = irc.IsChannel(msg)
	serviceMsg.Connection = conn.Name
	serviceMsg.Meta = &TenyksMeta{"Tenyks", TenyksVersion}
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
	meta := msg.Meta.(ServiceMeta)
	srv := &Service{}
	srv.Name = meta.Name
	srv.Version = meta.Version
	srv.Online = true
	srv.LastPing = time.Now()
	self.engine.ServiceRg.RegisterService(srv)
}

func (self *Connection) ByeServiceHandler(msg *Message) {
	meta := msg.Meta.(ServiceMeta)
	srv := self.engine.ServiceRg.GetServiceByName(meta.Name)
	if srv != nil {
		srv.Online = false
	}
}
