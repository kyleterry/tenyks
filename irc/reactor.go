package irc

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
					log.Error("[%s] Could not connect.", conn.Name)
					break
				}
			}
			select {
			case rawmsg := <-conn.In:
				msg := ParseMessage(rawmsg)
				msg.Conn = conn
				if msg != nil { // Just ignore invalid messages. Who knows...
					dispatch(msg.Command, conn, msg)
				}
			case <-reactorCtl:
				break
			}
		}
	} else {
		log.Error("[%s] Could not connect.", conn.Name)
	}
}

func dispatch(command string, conn *Connection, msg *Message) {
	handlers, ok := conn.Registry.handlers[command]
	if ok {
		log.Debug("[%s] Dispatching handler `%s`", conn.Name, command)
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(fn)
			go handler(conn, msg)
		}
	}
}
