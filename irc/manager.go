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
	RtSetNick
	ReOK
	ReErr
)

type Request struct {
	RequestType int
	Requirement int
	ConnName string
	Payload interface{}
	Response Response
}

type Response struct {
	ResponseType int
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

func StartConnection(conn *Connection) {
	wait := conn.Connect()
	<-wait
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
						if req.Requirement == RqResponseRequired {
							req.Response = Response{
								ResponseType: ReOK,
								Payload: "Connection was added and is boostrapping now",
							}
							com <- req
							go StartConnection(conn)
						}
					case RtDisconnect:
						conn := cm.ConnFromName(req.ConnName)
						if conn != nil {
							conn.Disconnect()
						}
					case RtReconnect:
						conn := cm.ConnFromName(req.ConnName)
						if conn != nil {
							conn.Connect()
						}
					case RtJoin:
						conn := cm.ConnFromName(req.ConnName)
						if conn != nil {
							conn.JoinChannel(req.Payload.(string))
						}
					case RtPart:
						conn := cm.ConnFromName(req.ConnName)
						if conn != nil {
							conn.PartChannel(req.Payload.(string))
						}
					case RtSetNick:
						conn := cm.ConnFromName(req.ConnName)
						if conn != nil {
							conn.SetNick(req.Payload.(string))
						}
					}
				case <-done:
				case <-time.After(time.Second * 2):
				}
			case <-done:
				return
			}
		}
	}()

	return c
}

func (cm *ConnectionManager) ConnFromName(name string) *Connection {
	if conn, ok := cm.connections[name]; ok {
		return conn
	}
	return nil
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
