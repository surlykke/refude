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
	"time"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var notificationExpireryHints = make(chan struct{})
var flashExpireryHints = make(chan struct{})

func Run() {
	go DoDBus()

	for {
		select {
		case notification := <-incomingNotifications:
			putNotification(notification)
		case rem := <-removals:
			removeNotification(rem.id, rem.reason)
		case <-notificationExpireryHints:
			removeExpired()
		case <-flashExpireryHints:
			removeExpiredFlash()
		}
	}
}

func putNotification(notification *Notification) {
	var path = fmt.Sprintf("/notification/%X", notification.Id)
	var res = resource.MakeResource(path, notification.Subject, notification.Body, notification.iconName, "notification", notification)

	var currentFlash = getFlash()
	if currentFlash == nil || currentFlash.Urgency <= notification.Urgency {
		setFlash(res)
		if notification.Urgency == low {
			time.AfterFunc(time.Millisecond*2050, func() { flashExpireryHints <- struct{}{} })
		} else if notification.Urgency == normal {
			time.AfterFunc(time.Millisecond*6050, func() { flashExpireryHints <- struct{}{} })
		}
		watch.SomethingChanged("/notification/flash")
	}

	Notifications.Put(res)
}

func removeExpired() {
	fmt.Println("removeExpired")
	var somethingExpired = false
	for _, res := range Notifications.GetAll() {
		var notification = res.Data.(*Notification)
		if notification.Urgency < critical && notification.Expires.Before(time.Now()) {
			Notifications.Delete(fmt.Sprintf("/notification/%X", notification.Id))
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", notification.Id, Expired)
			somethingExpired = true
		}
	}

	if somethingExpired {
		somethingChanged()
	}
}

func removeExpiredFlash() {
	var currentFlash = getFlash()
	if currentFlash != nil {
		if currentFlash.Urgency == normal && currentFlash.Created.Before(time.Now().Add(-6*time.Second)) ||
			currentFlash.Urgency == low && currentFlash.Created.Before(time.Now().Add(-2*time.Second)) {

			setFlash(nil)
			watch.SomethingChanged("/notification/flash")
		}
	}

}

func removeNotification(id uint32, reason uint32) {
	fmt.Println("In removeNotification")
	var path = fmt.Sprintf("/notification/%X", id)
	if found := Notifications.Delete(path); found {
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		fmt.Println("somethingChanged...")
		somethingChanged()
	}

	var currentFlash = getFlash()
	if currentFlash != nil && currentFlash.Id == id {
		setFlash(nil)
		watch.SomethingChanged("/notification/flash")
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notification/flash" {
		var res = getFlashResource()
		if res == nil {
			respond.NotFound(w)
		} else {
			res.ServeHTTP(w, r)
		}
	} else {
		Notifications.ServeHTTP(w, r)
	}
}
