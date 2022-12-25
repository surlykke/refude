// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
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


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notification/flash" {
		if r.Method == "GET" {
			if flash := getFlash(); flash != nil {
				respond.AsJson(w, resource.MakeWrapper[uint32]("/notification/self", flash, ""))
			} else {
				respond.NotFound(w)
			}
		} else {
			respond.NotAllowed(w)
		}
	} else {
		Notifications.ServeHTTP(w, r)
	}
}

func getFlash() *Notification{
	var notifications = Notifications.GetAll()
	for i,j := 0, len(notifications) - 1; i < j; i, j = i+1, j-1 {
		notifications[i],notifications[j] = notifications[j],notifications[i]
	}
	var now = time.Now().UnixMilli()
	for _, n := range notifications {
		if n.Urgency == Critical && now < n.Created + 3600000 {
			return n
		}
	} 
	for _, n := range notifications {
		if n.Urgency == Normal && now < n.Created + 10000 {
			return n
		}
	} 
	for _, n := range notifications {
		if n.Urgency == Low && now < n.Created + 4000 {
			return n
		}
	}
	return nil
}



