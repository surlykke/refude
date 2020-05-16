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
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications/osd"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/watch"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notification/osd" {
		var current = osd.CurrentlyShowing()
		if current != nil {
			respond.AsJson(w, r, current)
		} else {
			respond.Ok(w)
		}
	} else if r.URL.Path == "/notifications" {
		respond.AsJson(w, r, Collect(searchutils.Term(r)))
	} else if notification := getNotification(r.URL.Path); notification != nil {
		if r.Method == "POST" && notification.haveDefaultAction() {
			respond.AcceptedAndThen(w, func() {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", notification.Id, "default")
			})
		} else if r.Method == "DELETE" {
			respond.AcceptedAndThen(w, func() { removals <- removal{id: notification.Id, reason: Dismissed} })
		} else {
			respond.AsJson(w, r, notification.ToStandardFormat())
		}
	} else {
		respond.NotFound(w)
	}
}

func Collect(term string) respond.StandardFormatList {
	lock.Lock()
	defer lock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(notifications))
	for _, notification := range notifications {
		if rank := searchutils.SimpleRank(notification.Subject, notification.Body, term); rank > -1 {
			sfl = append(sfl, notification.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl
}

// Notifications that have a default action
func CollectActionable(term string) respond.StandardFormatList {
	lock.Lock()
	defer lock.Unlock()
	var sfl = make(respond.StandardFormatList, 0, len(notifications))
	for _, notification := range notifications {
		if _, ok := notification.Actions["default"]; ok {
			if rank := searchutils.SimpleRank(notification.Subject, notification.Body, term); rank > -1 {
				sfl = append(sfl, notification.ToStandardFormat().Ranked(rank))
			}
		}
	}
	return sfl
}

func AllPaths() []string {
	lock.Lock()
	defer lock.Unlock()
	var paths = make([]string, 0, len(notifications)+2)
	for _, n := range notifications {
		paths = append(paths, n.path)
	}
	paths = append(paths, "/notifications")
	paths = append(paths, "/notification/osd")
	return paths
}

var lock sync.Mutex
var notifications = []*Notification{}

func getNotification(path string) *Notification {
	lock.Lock()
	defer lock.Unlock()
	for _, notification := range notifications {
		if notification.path == path {
			return notification
		}
	}
	return nil
}

func upsert(notification *Notification) {
	lock.Lock()
	defer lock.Unlock()
	for i := 0; i < len(notifications); i++ {
		if notifications[i].Id == notification.Id {
			notifications[i] = notification
			return
		}
	}
	notifications = append(notifications, notification)
}

func sendToOsd(n *Notification) {
	if categoryHint, ok := n.Hints["category"]; ok {
		if category, ok := categoryHint.(string); ok {
			if strings.HasPrefix(category, "x-org.refude.gauge.") {
				if tmp, err := strconv.Atoi(n.Body); err != nil {
					fmt.Println("Error converting body to int: ", err)
				} else if tmp < 0 || tmp > 100 {
					fmt.Println("gauge not in acceptable range:", tmp)
				} else {
					osd.PublishGauge(n.Id, n.Sender, n.IconName, uint8(tmp))
					return
				}
			}
		}
	}
	osd.PublishMessage(n.Id, n.Sender, n.Subject, n.Body, n.IconName)
}

func removeNotification(id uint32) *Notification {
	lock.Lock()
	defer lock.Unlock()
	for i := 0; i < len(notifications); i++ {
		if notifications[i].Id == id {
			var notification = notifications[i]
			notifications = append(notifications[:i], notifications[i+1:]...)
			return notification
		}
	}
	return nil
}

/**
 * Returns:
 * if not found: nil, false
 * if found, but not expired, notification, false, and notification is not removed
 * if found and expired, notification, true, and notification is removed
 */
func removeIfExpired(id uint32) (*Notification, bool) {
	lock.Lock()
	defer lock.Unlock()
	for i := 0; i < len(notifications); i++ {
		if notifications[i].Id == id {
			var notification = notifications[i]
			if notification.Expires.Before(time.Now()) {
				notifications = append(notifications[:i], notifications[i+1:]...)
				return notification, true
			} else {
				return notification, false
			}
		}
	}
	return nil, false
}

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var reaper = make(chan uint32)

func Run() {
	go osd.RunOSD()
	go DoDBus()

	for {
		select {
		case notification := <-incomingNotifications:
			upsert(notification)
			sendToOsd(notification)
			sendEvent(notification.path)
		case rem := <-removals:
			if notification := removeNotification(rem.id); notification != nil {
				sendEvent(notification.path)
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", rem.id, rem.reason)
			}
		case id := <-reaper:
			if notification, wasExpired := removeIfExpired(id); notification != nil {
				if wasExpired {
					sendEvent(notificationSelf(id))
					conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, Expired)
				} else {
					time.AfterFunc(notification.Expires.Sub(time.Now())+100*time.Millisecond, func() {
						reaper <- notification.Id
					})
				}
			}
		}
	}
}

func sendEvent(path string) {
	watch.SomethingChanged(path)
}
