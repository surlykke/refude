// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"strconv"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/x11"
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
				Notifications.Delete(notification.GetPath())
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.NotificationId, Expired)
				count++
			}
		}
	}
	if count > 0 {
		watch.NotificationChanged()
	}
}

func removeNotification(id uint32, reason uint32) {
	if deleted := Notifications.Delete(strconv.Itoa(int(id))); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		watch.NotificationChanged()
	}
}

func GetFlash(string) resource.Resource {
	var notifications = Notifications.GetAll()
	var now = time.Now().UnixMilli()
	for _, urgency := range []Urgency{Critical, Normal, Low} {
		for _, notification := range notifications {
			if notification.Urgency == urgency && !timedOut(notification, now) {
				return notification
			}
		}
	}
	
	return nil
}

func timedOut(flash *Notification, now int64) bool {
	if flash.Urgency == Critical {
		return now > flash.Created+3600000
	} else if flash.Urgency == Normal {
		return now > flash.Created+10000
	} else { // Low
		return now > flash.Created+4000
	}
}


func notifierShow() {
	if !x11.PurgeAndShow("localhost__refude_html_notifier", false) {
		xdg.RunCmd(xdg.BrowserCommand, "--app=http://localhost:7938/refude/html/notifier")
	}
}

