package irc

import (
	"time"
)

const (
	RqNoResponse = iota
	RqResponseRequired
	RtJoin
	RtPart
	RtInfo
	RtNewConn
	RtDisconnect
	RtReconnect
)

type Request struct {
	RequestType int
	Requirement int
	Payload interface{}
}

type ConnectionManager struct {
	connections map[string]*Connection
	cmCom chan chan Request
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
	}
}

func (cm *ConnectionManager) Start(done chan bool) chan chan Request {
	c := make(chan chan Request)

	// Event loop
	go func() {
		for {
			select {
			case com := <-c:
				select {
				case req := <-com:
					if req.Requirement == RqNoResponse {
						close(com)
					}
					switch req.RequestType {
					case RtNewConn:
						conn := req.Payload.(*Connection)
						cm.connections[conn.Name] = conn
					case RtDisconnect:
						connName := req.Payload.(string)
						if conn, ok := cm.connections[connName]; ok {
							conn.Disconnect()
						}
					}
				case <-done:
				case <-time.After(time.Second * 5):
				}
			case <-done:
				return
			}
		}
	}()

	return c
}

func (cm *ConnectionManager) AddConnection(conn *Connection) {
	r := Request{
		RequestType: RtNewConn,
		Requirement: RqNoResponse,
	}
	c := make(chan Request)
	c <- r
	cm.cmCom <- c
	<-c
}
