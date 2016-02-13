package irc

type IRCConnections map[string]*Connection

func ConnectionReactor(conn *Connection, reactorCtl <-chan bool) {
	log.Info("[%s] Connecting...", conn.Name)
	connected := <-conn.Connect()
	log.Info("[%s] Connected!", conn.Name)
	dispatch("bootstrap", conn, nil)
	if connected == true {
		for {
			if conn.IsConnected() == false {
				connected := <-conn.Connect()
				if connected == false {
					log.Errorf("[%s] Could not connect.", conn.Name)
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
		log.Errorf("[%s] Could not connect.", conn.Name)
	}
}

func dispatch(command string, conn *Connection, msg *Message) {
	conn.Registry.RegistryMu.Lock()
	defer conn.Registry.RegistryMu.Unlock()
	handlers, ok := conn.Registry.Handlers[command]
	if ok {
		log.Debug("[%s] Dispatching handler `%s`", conn.Name, command)
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(*Handler)
			go handler.Fn(conn, msg)
		}
	}
}
