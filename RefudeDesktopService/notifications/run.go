// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
)


var removals = make(chan removal)

func Run(notificationsCollection *NotificationsCollection) {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			notificationsCollection.mutex.Lock()
			notificationsCollection.notifications[notification.Self] = notification
			notificationsCollection.CachingJsonGetter.ClearByPrefixes(fmt.Sprintf("/notification/%d", notification.Id), "/notifications")
			notificationsCollection.mutex.Unlock()
		case rem := <-removals:
			fmt.Println("Got removal..")
			notificationsCollection.mutex.Lock()
			if notification, ok := notificationsCollection.notifications[notificationSelf(rem.id)]; ok {
				if rem.internalId == 0 || rem.internalId == notification.internalId {
					//resourceMap.Unmap(resource.Standardizef("/notifications/%d", rem.id))
					delete(notificationsCollection.notifications, notification.Self)
					notificationClosed(rem.id, rem.reason)
					notificationsCollection.CachingJsonGetter.ClearByPrefixes(fmt.Sprintf("/notification/%d", rem.id), "/notifications")

				}
			}
			notificationsCollection.mutex.Unlock()
		}
	}
}


