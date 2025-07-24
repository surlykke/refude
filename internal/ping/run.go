// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
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
