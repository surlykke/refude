// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"sort"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var notifications = make(map[uint32]*Notification)

var notificationsMap = resource.MakeResourceMap()
var Notifications = resource.MakeServer(notificationsMap)

var removals = make(chan removal)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	updateCollections()
	for {
		select {
		case notification := <-updates:
			notifications[notification.Id] = notification
			notificationsMap.Set(notificationSelf(notification.Id), resource.MakeJsonResouceWithEtag(notification))
			updateCollections()
		case rem := <-removals:
			var path = string(notificationSelf(rem.id))
			if notification, ok := notifications[rem.id]; !ok {
				continue
			} else if rem.reason == Expired && !notificationIsExpired(notification) {
				continue
			} else {
				delete(notifications, rem.id)
				notificationsMap.Remove(path)
				updateCollections()
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		}
		notificationsMap.Broadcast()
	}
}

func updateCollections() {
	var lst = make(resource.Selfielist, 0, len(notifications))
	for _, notification := range notifications {
		lst = append(lst, notification)
	}
	sort.Sort(lst)
	notificationsMap.Set("/notifications", resource.MakeJsonResouceWithEtag(lst))
	notificationsMap.Set("/notifications/brief", resource.MakeJsonResouceWithEtag(lst.GetSelfs()))
}

func notificationIsExpired(res interface{}) bool {
	n, ok := res.(*Notification)
	if !ok {
		return false
	}
	return !time.Now().Before(n.Expires)
}
