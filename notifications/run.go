// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/watch"
)

var notificationExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()
	for range time.NewTicker(30 * time.Minute).C {
		removeExpired()
	}

}

func removeExpired() {
	var count = 0
	for _, notification := range Notifications.GetAll() {
		if notification.Urgency < Critical {
			if notification.Expires < time.Now().UnixMilli() {
				Notifications.Delete(notification.NotificationId)
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.NotificationId, Expired)
				count++
			}
		}
	}
	if count > 0 {
		updateFlash()
		watch.SomethingChanged("/notification/")
	}
}

func removeNotification(id uint32, reason uint32) {
	if deleted := Notifications.Delete(id); deleted {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		updateFlash()
		watch.SomethingChanged("/notification/")
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notification/flash" {
		if r.Method == "GET" {
			flashMutex.Lock()
			var flashCopy = flash
			flashMutex.Unlock()

			if flashCopy != nil {
				respond.AsJson(w, resource.MakeWrapper[uint32](fmt.Sprintf("/notification/%d", flashCopy.Id()), flashCopy, ""))
			} else {
				respond.NotFound(w)
			}
		} else {
			respond.NotAllowed(w)
		}
	} else {
		Notifications.ServeHTTP(w, r)
	}
}

var flash *Notification = nil
var flashMutex sync.Mutex

func updateFlash() {
	flashMutex.Lock()
	defer flashMutex.Unlock()

	var newFlash = getFlash()

	if newFlash == nil && flash != nil {
		notifierHide()
	} else if newFlash != nil && flash == nil {
		notifierShow()
	}

	if newFlash != nil && newFlash != flash {
		var timeout time.Time
		var created = time.UnixMilli(newFlash.Created)
		switch newFlash.Urgency {
		case Critical:
			timeout = created.Add(1 * time.Hour)
		case Normal:
			timeout = created.Add(10 * time.Second)
		case Low:
			timeout = created.Add(4 * time.Second)
		default:
			timeout = time.Now()
		}
		timeout = timeout.Add(50 * time.Millisecond)

		time.AfterFunc(timeout.Sub(time.Now()), updateFlash)
	}

	flash = newFlash

	watch.SomethingChanged("/notification/flash")
}

func getFlash() *Notification {
	var newFlash *Notification = nil

	var now = time.Now().UnixMilli()
	for _, n := range Notifications.GetAll() {
		if !timedOut(n, now) && !shadedBy(n, newFlash) {
			newFlash = n
		}
	}

	return newFlash
}

func timedOut(flash *Notification, now int64) bool {
	if flash.Urgency == Critical {
		return now > flash.Created+3600000
	} else if flash.Urgency == Normal {
		return now > flash.Created+10000
	} else { // Low
		return now > flash.Created+4000
	}
}

func shadedBy(flash, other *Notification) bool {
	return other != nil && other.Urgency > flash.Urgency
}

func notifierShow() {
	xdg.RunCmd("notifierShow")
}

func notifierHide() {
	xdg.RunCmd("notifierHide")
}
