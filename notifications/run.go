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
	go maintainFlash()
	for range time.NewTicker(30 * time.Minute).C {
		removeExpired()
	}

}

func removeExpired() {
	var count = 0
	for _, notification := range Notifications.GetAll() {
		if notification.Urgency < Critical {
			if time.Time(notification.Expires).Before(time.Now()) {
				Notifications.Delete(notification.GetId())
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.NotificationId, Expired)
				count++
			}
		}
	}
	if count > 0 {
		flashPing <- struct{}{}
	}
}

func removeNotification(id uint32, reason uint32) {
	if deleted := Notifications.Delete(strconv.Itoa(int(id))); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		flashPing <- struct{}{}
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


var flashPing = make(chan struct{})

var flash *Notification

const flashTimeoutLow time.Duration = 2 * time.Second
const flashTimeoutNormal time.Duration = 6 * time.Second
const _50ms = 50 * time.Millisecond

func maintainFlash() {
	for range flashPing {
		var critical, normal, low *Notification
		var now = time.Now()
		for _, n := range Notifications.GetAll() {
			if n.Urgency == Critical {
				critical = n 
				break
			} else if n.Urgency == Normal {
				if normal == nil && now.Before(time.Time(n.Created).Add(flashTimeoutNormal)) {
					normal = n
				}
			} else {
				if low == nil && now.Before(time.Time(n.Created).Add(flashTimeoutLow)) {
					low = n
				}
			}
		}
	
		if critical != nil {
			flash = critical
		} else if normal != nil {
			flash = normal
		} else {
			flash = low 
		}

		if flash != nil {
			if flash.Urgency == Normal {
				time.AfterFunc(time.Time(flash.Created).Add(flashTimeoutNormal).Sub(now), func() {flashPing <- struct{}{}})
			} else if flash.Urgency == Low {
				time.AfterFunc(time.Time(flash.Created).Add(flashTimeoutLow).Sub(now), func() {flashPing <- struct{}{}})
			}
		}

		fmt.Println("Flash now:", flash)
	}
		
}
