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

var notificationsMap = resource.MakeResourceMap("/notifications")
var Notifications = resource.MakeJsonResourceServer(notificationsMap)

var removals = make(chan removal)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			notificationsMap.Set(notificationSelf(notification.Id), notification)
		case rem := <-removals:
			var path = string(notificationSelf(rem.id))
			if rem.reason == Expired && notificationsMap.RemoveIf(path, notificationIsExpired) ||
				rem.reason == Dismissed && notificationsMap.Remove(path) ||
				rem.reason == Closed && notificationsMap.Remove(path) {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		}

	}
}

func notificationIsExpired(res interface{}) bool {
	n, ok := res.(*Notification)
	if !ok {
		return false
	}
	return !time.Now().Before(n.Expires)
}
