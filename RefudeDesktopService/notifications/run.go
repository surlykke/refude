// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

var removals = make(chan removal)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			setNotification(notification)
		case rem := <-removals:
			if removeNotification(notificationSelf(rem.id), rem.internalId) {
				notificationClosed(rem.id, rem.reason)
			}
		}
	}
}
