package service

import (
	"time"
	"encoding/json"
)

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
	self.engine.UpdateService(meta.Name, ServiceOnline)
}
