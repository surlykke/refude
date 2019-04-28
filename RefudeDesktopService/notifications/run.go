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

var Notifications = func() *resource.GenericResourceCollection {
	var grc = resource.MakeGenericResourceCollection()
	grc.AddCollectionResource("/notifications", "/notification/")
	return grc
}()

var removals = make(chan removal)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			Notifications.Set(notification)
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
	n, ok := res.(*Notification)
	return ok && !time.Now().Before(n.Expires)
}
