package ping

import (
	"golang.org/x/net/websocket"
)

var WebsocketHandler = websocket.Handler(func(conn *websocket.Conn) {
	var bytes = make([]byte, 100)
	var err error 
		
	for {
		if _, err = conn.Read(bytes); err != nil {
			conn.Close()
			return
		} 
	}
})

