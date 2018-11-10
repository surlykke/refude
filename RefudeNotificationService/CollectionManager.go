// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/resource"
)

var removals = make(chan removal)
var notifications = make(map[uint32]*Notification)

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			resourceMap.Unmap(resource.Standardizef("/notifications/%d", notification.Id))
			resourceMap.RemoveAll(resource.Standardizef("/actions/%d", notification.Id))
			notifications[notification.Id] = notification
			resourceMap.Map(notification)
		case rem := <-removals:
			fmt.Println("Got removal..")
			if notification, ok := notifications[rem.id]; ok {
				if rem.internalId == 0 || rem.internalId == notification.internalId {
					resourceMap.Unmap(resource.Standardizef("/notifications/%d", rem.id))
					resourceMap.RemoveAll(resource.Standardizef("/actions/%d", rem.id))
					delete(notifications, rem.id)
					notificationClosed(rem.id, rem.reason)
				}
			}
		}
	}
}


