package irc

func ConnectionReactor(conn *Connection, reactorCtl <-chan bool) {
	log.Info("[%s] Connecting", conn.Name)
	connected := <-conn.Connect()
	if connected == true {
		Bootstrap(conn)
		for {
			if conn.IsConnected() == false {
				connected := <-conn.Connect()
				if connected == true {
					Bootstrap(conn)
				} else {
					log.Error("[%s] Could not connect.", conn.Name)
					break
				}
			}
			select {
			case rawmsg := <-conn.In:
				msg := ParseMessage(rawmsg)
				log.Debug("%+v", *msg)
				if msg != nil { // Just ignore invalid messages. Who knows...
					dispatch(conn, msg)
				}
			case <-reactorCtl:
				break
			}
		}
	} else {
		log.Error("[%s] Could not connect.", conn.Name)
	}
}

func dispatch(conn *Connection, msg *Message) {
	handlers := ircHandlers.handlers[msg.Command]
	for i := handlers.Front(); i != nil; i = i.Next() {
		i.Value.()
	}
}
