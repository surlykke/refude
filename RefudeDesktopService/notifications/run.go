// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"net/http"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/ss_events"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if notification := getNotification(r.URL.Path); notification != nil {
		if r.Method == "GET" {
			respond.AsJson(w, notification.ToStandardFormat())
		} else if r.Method == "POST" && notification.haveDefaultAction() {
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", notification.Id, "default")
			respond.Accepted(w)
		} else if r.Method == "DELETE" {
			removals <- removal{id: notification.Id, reason: Dismissed}
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	} else if notificationImage := getNotificationImage(r.URL.Path); notificationImage != nil {
		if r.Method == "GET" {
			http.ServeFile(w, r, notificationImage.imagePath)
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}

}

func SearchNotifications(collector *searchutils.Collector) {
	lock.Lock()
	defer lock.Unlock()

	for _, notification := range notifications {
		collector.Collect(notification.ToStandardFormat())
	}
}

func AllPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(notifications))
	for path, _ := range notifications {
		paths = append(paths, path)
	}
	return paths
}

var lock sync.Mutex
var notifications = make(map[string]*Notification)
var notificationImages = make(map[string]*NotificationImage)

func getNotification(path string) *Notification {
	lock.Lock()
	defer lock.Unlock()
	return notifications[path]
}

func setNotification(notification *Notification) {
	lock.Lock()
	defer lock.Unlock()
	notifications[notificationSelf(notification.Id)] = notification
}

func getNotificationImage(path string) *NotificationImage {
	lock.Lock()
	defer lock.Unlock()
	return notificationImages[path]
}

func setNotificationImage(notificationId uint32, notificationImage *NotificationImage) {
	lock.Lock()
	defer lock.Unlock()
	notificationImages[notificationImageSelf(notificationId)] = notificationImage
}

func removeNotification(id uint32) bool {
	lock.Lock()
	defer lock.Unlock()
	var self = notificationSelf(id)
	_, ok := notifications[self]
	if ok {
		delete(notifications, self)
		delete(notificationImages, self)
	}
	return ok
}

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var reaper = make(chan uint32)

func Run() {
	go DoDBus()

	for {
		select {
		case notification := <-incomingNotifications:
			var self = notificationSelf(notification.Id)
			if "" != notification.imagePath {
				notification.Image = notificationImageSelf(notification.Id)
			}

			lock.Lock()
			notifications[self] = notification
			if notification.Image != "" {
				notificationImages[notification.Image] = &NotificationImage{notification.imagePath}
			}
			lock.Unlock()
			sendEvent(self)
		case rem := <-removals:
			var self = notificationSelf(rem.id)
			var imageSelf = notificationImageSelf(rem.id)
			var found bool

			lock.Lock()
			if _, found = notifications[self]; found {
				delete(notifications, self)
				delete(notificationImages, imageSelf)
			}
			lock.Unlock()
			if found {
				sendEvent(self)
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		case id := <-reaper:
			var n *Notification
			var self = notificationSelf(id)
			var imageSelf = notificationImageSelf(id)
			lock.Lock()
			n = notifications[self]
			lock.Unlock()

			if n != nil {
				var now = time.Now()
				if now.Before(n.Expires) {
					time.AfterFunc(n.Expires.Sub(now)+100*time.Millisecond, func() { reaper <- n.Id })
				} else {
					lock.Lock()
					delete(notifications, self)
					delete(notificationImages, imageSelf)
					lock.Unlock()
					sendEvent(self)
					conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, Expired)
				}
			}
		}
	}
}

func sendEvent(path string) {
	ss_events.Publish <- &ss_events.Event{Type: "notification", Path: path}
}
