// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"time"
)

var notificationExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()
	for range time.NewTicker(30 * time.Minute).C {
		removeExpired()
	}

}

func removeExpired() {
	var somethingExpired = false
	for _, res := range Notifications.GetAll() {
		var notification = res.(*Notification)
		if notification.Urgency < Critical {
			if notification.Expires < time.Now().UnixMilli() {
				Notifications.Delete(fmt.Sprintf("/notification/%d", notification.Id))
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.Id, Expired)
				somethingExpired = true
			}
		}
	}

	if somethingExpired {
		somethingChanged()
	}
}

func removeNotification(id uint32, reason uint32) {
	var path = fmt.Sprintf("/notification/%d", id)
	if deleted := Notifications.Delete(path); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		somethingChanged()
	}
}
