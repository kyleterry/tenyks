package irc

func ConnectionReactor(conn *Connection, reactorCtl <-chan bool) {
	log.Info("[%s] Connecting", conn.Name)
	connected := <-conn.Connect()
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
				log.Debug("%+v", *msg)
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
	log.Debug("%s\n", ok)
	if ok {
		for i := handlers.Front(); i != nil; i = i.Next() {
			handler := i.Value.(fn)
			handler(conn, msg)
		}
	}
}
