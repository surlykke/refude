// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"net/http"
	"time"

	"github.com/surlykke/RefudeServices/lib/respond"
)

var notificationExpireryHints = make(chan struct{})
var flashExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()
}

func removeExpired() {
	var somethingExpired = false
	for _, res := range Notifications.GetAll() {
		var notification = res.Data.(*Notification)
		if notification.Urgency < Critical {
			if notification.Expires.Before(time.Now()) {
				Notifications.Delete(fmt.Sprintf("/notification/%X", notification.Id))
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
	var path = fmt.Sprintf("/notification/%X", id)
	if deleted := Notifications.Delete(path); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		somethingChanged()
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notification/flash" {
		if res := getFlashResource(); res == nil {
			respond.NotFound(w)
		} else {
			res.ServeHTTP(w, r)
		}
	} else {
		Notifications.ServeHTTP(w, r)
	}
}
