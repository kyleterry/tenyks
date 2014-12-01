package control

import (
	"bufio"
	"io"
	"errors"
	"fmt"
	"github.com/kyleterry/tenyks/config"
	"github.com/kyleterry/tenyks/irc"
	"github.com/op/go-logging"
	"net"
)

var log = logging.MustGetLogger("tenyks")

type ControlServer struct {
	socket   net.Listener
	ctl      chan bool
	ircconns *irc.IRCConnections
	config   config.ControlConfig
}

type ControlConnection struct {
	conn net.Conn
	io   *bufio.ReadWriter
}

func NewControlServer(conf config.ControlConfig) (*ControlServer, error) {
	if conf.Bind == "" {
		return nil, errors.New("Control server config needs a bind setting")
	}
	cs := &ControlServer{}
	cs.ctl = make(chan bool, 1)
	cs.config = conf
	return cs, nil
}

func (serv *ControlServer) SetIRCConns(ircconns *irc.IRCConnections) {
	serv.ircconns = ircconns
}

func (serv *ControlServer) Start() (chan bool, error) {
	wait := make(chan bool, 1)
	sock, err := net.Listen("tcp", serv.config.Bind)
	if err != nil {
		return nil, err
	}
	serv.socket = sock

	go func() {
		defer close(wait)
		accept := func() <-chan ControlConnection {
			a := make(chan ControlConnection)
			go func() {
				for {
					conn, err := serv.socket.Accept()
					if err != nil {
						log.Error("Error while accepting connection")
					}
					a <- ControlConnection{conn, bufio.NewReadWriter(
						bufio.NewReader(conn),
						bufio.NewWriter(conn),
					)}
				}
			}()
			return a
		}()

		wait <- true

		for {
			select {
			case conn := <-accept:
				go serv.connectionWorker(conn)
			case <-serv.ctl:
				return
			}
		}

	}()

	return wait, nil
}

func (serv *ControlServer) Stop() error {
	serv.ctl <- true
	close(serv.ctl)
	err := serv.socket.Close()
	if err != nil {
		return err
	}
	return nil
}

func (serv *ControlServer) connectionWorker(controlConn ControlConnection) {
	defer controlConn.conn.Close()
	for {
		msg, err := controlConn.io.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Info("Client connection closed")
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				log.Error("Client timeout")
				return
			}
			log.Error("Could not read string from connection")
		}
		serv.handleMessage(msg)
	}
}

func (serv ControlServer) handleMessage(msg string) {
	fmt.Println(msg)
}
