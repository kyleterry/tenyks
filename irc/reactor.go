package irc

type IRCConnections map[string]*Connection

func ConnectionReactor(conn *Connection, reactorCtl <-chan bool) {
	Logger.Info("connecting to network", "connection", conn.Name)
	connected := <-conn.Connect()
	Logger.Info("connected successfully", "connection", conn.Name)
	dispatch("bootstrap", conn, nil)
	if connected == true {
		for {
			if conn.IsConnected() == false {
				connected := <-conn.Connect()
				if connected == false {
					Logger.Error("failed to connect", "connection", conn.Name)
					break
				}
				dispatch("bootstrap", conn, nil)
			}
			select {
			case rawmsg, ok := <-conn.In:
				if !ok { // Conn closed the channel because of a disconnect.
					continue
				}
				msg := ParseMessage(rawmsg)
				if msg != nil { // Just ignore invalid messages. Who knows...
					msg.Conn = conn
					dispatch(msg.Command, conn, msg)
				}
			case <-reactorCtl:
				break
			}
		}
	} else {
		Logger.Error("failed to connect", "connection", conn.Name)
	}
}

func dispatch(command string, conn *Connection, msg *Message) {
	conn.Registry.RegistryMu.Lock()
	defer conn.Registry.RegistryMu.Unlock()
	handlers, ok := conn.Registry.Handlers[command]
	if ok {
		Logger.Debug("dispatching handler", "connection", conn.Name, "command", command)
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(*Handler)
			go handler.Fn(conn, msg)
		}
	}
}
