// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"sort"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var notifications = make(map[uint32]*Notification)
var notificationImages = make(map[string]*NotificationImage)

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var reaper = make(chan uint32)

func Run() {
	go DoDBus()

	updateCollections()
	for {
		select {
		case notification := <-incomingNotifications:
			fmt.Println("Notification")
			notifications[notification.Id] = notification
			var notificationImagePath = fmt.Sprintf("/notificationimage/%d", notification.Id)
			if "" != notification.imagePath {
				notification.Image = notificationImagePath
				notificationImages[notificationImagePath] = &NotificationImage{notification.imagePath}
			}
			updateCollections()
		case rem := <-removals:
			if n, ok := notifications[rem.id]; !ok {
				continue
			} else {
				delete(notifications, rem.id)
				delete(notificationImages, n.Image)
				updateCollections()
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		case id := <-reaper:
			if n, ok := notifications[id]; ok {
				var now = time.Now()
				var age = now.Sub(n.Created)
				if age < time.Hour {
					time.AfterFunc(time.Minute*61-age, func() {
						reaper <- n.Id
					})
				} else {
					delete(notifications, id)
					delete(notificationImages, n.Image)
					conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, Expired)
				}
				updateCollections()
			}
		}
	}
}

func updateCollections() {
	var resources = make(map[string]resource.Resource)
	for path, notificationImage := range notificationImages {
		resources[path] = notificationImage
	}
	var notificationList = make(resource.ResourceList, 0, len(notifications))
	var notificationPaths = make(resource.BriefList, 0, len(notifications))
	for _, notification := range notifications {
		var path = notificationSelf(notification.Id)
		resources[path] = notification
		notificationList = append(notificationList, notification)
		notificationPaths = append(notificationPaths, path)
	}
	sort.Sort(notificationList)
	resources["/notifications"] = notificationList
	sort.Sort(notificationPaths)
	resources["/notificationpaths"] = notificationPaths

	resource.MapCollection(&resources, "notifications")
}

func notificationIsExpired(res interface{}) bool {
	n, ok := res.(*Notification)
	if !ok {
		return false
	}
	return !time.Now().Before(n.Expires)
}
