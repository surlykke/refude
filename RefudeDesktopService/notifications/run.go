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
	"strings"
)

var notificationCollection = MakeNotificationCollection()

var removals = make(chan removal)

func Serve(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/notification") {
		return false
	}

	if r.Method == "GET" {
		notificationCollection.GET(w, r)
	} else if r.Method == "POST" {
		notificationCollection.POST(w, r)
	} else if r.Method == "DELETE" {
		notificationCollection.DELETE(w, r)
	}

	return true
}

func Run() {
	var updates = make(chan *Notification)
	go DoDBus(updates, removals)

	for {
		select {
		case notification := <-updates:
			notificationCollection.mutex.Lock()
			notificationCollection.notifications[notification.GetSelf()] = notification
			notificationCollection.CachingJsonGetter.ClearByPrefixes(fmt.Sprintf("/notification/%d", notification.Id), "/notifications")
			notificationCollection.mutex.Unlock()
		case rem := <-removals:
			fmt.Println("Got removal..")
			notificationCollection.mutex.Lock()
			if notification, ok := notificationCollection.notifications[notificationSelf(rem.id)]; ok {
				if rem.internalId == 0 || rem.internalId == notification.internalId {
					//resourceMap.Unmap(resource.Standardizef("/notifications/%d", rem.id))
					delete(notificationCollection.notifications, notification.GetSelf())
					notificationClosed(rem.id, rem.reason)
					notificationCollection.CachingJsonGetter.ClearByPrefixes(fmt.Sprintf("/notification/%d", rem.id), "/notifications")

				}
			}
			notificationCollection.mutex.Unlock()
		}
	}
}


