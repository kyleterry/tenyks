package service

import (
	"encoding/json"

	"github.com/kyleterry/tenyks/irc"
	. "github.com/kyleterry/tenyks/version"
)

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

