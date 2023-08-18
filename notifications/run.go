// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"fmt"
	"strconv"
	"time"

	"github.com/surlykke/RefudeServices/lib/log"
	"golang.org/x/net/websocket"
)

var notificationExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()
	for range time.NewTicker(30 * time.Minute).C {
		removeExpired()
	}

}

func removeExpired() {
	var count = 0
	for _, notification := range Notifications.GetAll() {
		if notification.Urgency < Critical {
			if notification.Expires < time.Now().UnixMilli() {
				Notifications.Delete(notification.GetId())
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.NotificationId, Expired)
				count++
			}
		}
	}
}

func removeNotification(id uint32, reason uint32) {
	if deleted := Notifications.Delete(strconv.Itoa(int(id))); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
	}
}

var WebsocketHandler = websocket.Handler(func(conn *websocket.Conn) {
	var subscription = Notifications.Subscribe()
	for {
		subscription.Next()
		fmt.Println("Notification...")

		if err := websocket.JSON.Send(conn, Notifications.GetAll()); err != nil {
			log.Info(err)	
			return
		}
	}
})
