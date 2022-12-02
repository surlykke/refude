// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"time"

	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/watch"
)

var notificationExpireryHints = make(chan struct{})

func Run() {
	if config.NoNotificationServer() {
		return 
	}
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
				Notifications.Delete(notification.NotificationId)
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.NotificationId, Expired)
				count++	
			}
		}
	}
	if count > 0 {
		watch.SomethingChanged("/notification/")
	}
}

func removeNotification(id uint32, reason uint32) {
	if deleted := Notifications.Delete(id); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		watch.SomethingChanged("/notification/")
	}
}
