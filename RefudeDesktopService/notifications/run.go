// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var Notifications = resource.MakeGenericResourceCollection()

var removals = make(chan removal)

func Run() {
	Notifications.Set("/notifications", Notifications.MakePrefixCollection("/notification/"))
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			Notifications.Set(notificationSelf(notification.Id), resource.MakeJsonResource(notification))
		case rem := <-removals:
			var path = string(notificationSelf(rem.id))
			if rem.reason == Expired && Notifications.RemoveIf(path, notificationIsExpired) ||
				rem.reason == Dismissed && Notifications.Remove(path) ||
				rem.reason == Closed && Notifications.Remove(path) {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		}

	}
}

func notificationIsExpired(res resource.Resource) bool {
	jr, ok := res.(resource.JsonResource)
	if !ok {
		return false
	}
	n, ok := jr.Data.(*Notification)
	if !ok {
		return false
	}
	return !time.Now().Before(n.Expires)
}
