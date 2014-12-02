package irc

import (

)

const (
	RqNoResponse = iota
	RqResponseRequired
	RqJoin
	RqPart
	RqInfo
)

type Request struct {
	RequestType int
	Requirement int
}

type ConnectionManager struct {
	connections map[string]*Connection
}

func NewConnectionManager() *ConnectionManager {
	return &ConnectionManager{
		connections: make(map[string]*Connection),
	}
}

func (cm *ConnectionManager) Start(done chan bool) chan chan Event {
	c := make(chan chan Event)
	// Event loop
	go func() {
		for {
			select {
			case com := <-c:
				//handle
			case <-done:
				return
			}
		}
	}()
	return c
}

func (cm *ConnectionManager) HandleEvent(event Event) {

}
