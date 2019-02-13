// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/server"
)

var notificationsCollection = MakeNotificationsCollection()
var NotificationsServer = server.MakeServer(notificationsCollection)

var removals = make(chan removal)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			notificationsCollection.Lock()
			notificationsCollection.notifications[notification.Id] = notification
			notificationsCollection.JsonResponseCache.ClearByPrefixes(fmt.Sprintf("/notification/%d", notification.Id), "/notifications")
			notificationsCollection.Unlock()
		case rem := <-removals:
			fmt.Println("Got removal..")
			notificationsCollection.Lock()
			if notification, ok := notificationsCollection.notifications[rem.id]; ok {
				if rem.internalId == 0 || rem.internalId == notification.internalId {
					//resourceMap.Unmap(resource.Standardizef("/notifications/%d", rem.id))
					delete(notificationsCollection.notifications, rem.id)
					notificationClosed(rem.id, rem.reason)
					notificationsCollection.JsonResponseCache.ClearByPrefixes(fmt.Sprintf("/notification/%d", rem.id), "/notifications")

				}
			}
			notificationsCollection.Unlock()
		}
	}
}


