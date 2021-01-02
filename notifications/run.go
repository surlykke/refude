// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

var box, box2 atomic.Value

func init() {
	box.Store([]*Notification{})
	box2.Store(osdEvent{Type: none})
}

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var expireHints = make(chan struct{})
var scheduler = makeScheduler()

func Run() {
	go DoDBus()
	go scheduler.run()

	for {
		select {
		case notification := <-incomingNotifications:
			upsert(notification)
		case rem := <-removals:
			removeNotification(rem.id, rem.reason)
		case <-scheduler.ping:
			doMaintenance()
		}
	}
}

func getNotification(id uint32) *Notification {
	for _, notification := range box.Load().([]*Notification) {
		if notification.Id == id {
			return notification
		}
	}
	return nil
}

func upsert(notification *Notification) {
	var notifications = box.Load().([]*Notification)
	var next = make([]*Notification, len(notifications)+1, len(notifications)+1)
	var found = false
	for i, n := range notifications {
		if n.Id == notification.Id {
			next[i+1] = notification
			found = true
		} else {
			next[i+1] = n
		}
	}
	if found {
		next = next[1:]
	} else {
		next[0] = notification
	}
	box.Store(next)

	doMaintenance()
}

func removeNotification(id uint32, reason uint32) {
	var notifications = box.Load().([]*Notification)
	var next = make([]*Notification, 0, len(notifications))
	for _, n := range notifications {
		if n.Id == id {
			watch.SomethingChanged(n.self)
			watch.SomethingChanged("/notifications") // Should only happen once for the loop
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", n.Id, reason)
		} else {
			next = append(next, n)
		}
	}

	box.Store(next)
	doMaintenance()
}

func find(notifications []*Notification, id uint32) int {
	for i, n := range notifications {
		if n.Id == id {
			return i
		}
	}
	return -1
}

func doMaintenance() {
	removeExpired()

	var e1 = box2.Load().(osdEvent)
	var e2 = getOsdEvent()
	var same = e1.Type == e2.Type &&
		e1.Expires == e2.Expires &&
		e1.Sender == e2.Sender &&
		e1.Subject == e2.Subject &&
		len(e1.Body) == len(e2.Body) // oh well...

	if !same {
		if e2.Type != none {
			e2.Links = respond.Links{{
				Href:  "/notification/osd",
				Title: e2.Subject,

				Rel:     respond.Self,
				Profile: "/profile/osd",
			}}
			if e2.iconName != "" {
				e2.Links[0].Icon = icons.IconUrl(e2.iconName)
			} else if e2.Type == critical {
				e2.Links[0].Icon = icons.IconUrl("dialog-warning")
			} else if e2.Type == normal {
				e2.Links[0].Icon = icons.IconUrl("dialog-information")
			}
		}
		box2.Store(e2)
		watch.SomethingChanged("/notification/osd")
	}

	var nextMaintenance = time.Now().Add(1000 * time.Hour)
	if e2.Type != none {
		nextMaintenance = e2.Expires
	}
	for _, n := range box.Load().([]*Notification) {
		if n.Expires.Before(nextMaintenance) {
			nextMaintenance = n.Expires
		}
	}
	scheduler.schedule <- nextMaintenance.Add(100 * time.Millisecond)
}

func removeExpired() {
	var notifications = box.Load().([]*Notification)
	var now = time.Now()
	var haveRemovals bool
	var next = make([]*Notification, 0, len(notifications))
	for _, n := range notifications {
		if n.Expires.Before(now) {
			haveRemovals = true
			watch.SomethingChanged(n.self)
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", n.Id, Expired)
		} else {
			next = append(next, n)
		}
	}
	box.Store(next)
	if haveRemovals {
		watch.SomethingChanged("/notifications")
	}
}

func getOsdEvent() osdEvent {
	var now = time.Now()
	var gaugeTimeout = 2 * time.Second
	var normalTimeout = 6 * time.Second
	var urgentTimeout = 20 * time.Second
	var notifications = box.Load().([]*Notification)
	var event = osdEvent{Type: none}
	for _, n := range notifications {
		if n.isGauge() {
			if event.Type == none && now.Before(n.Created.Add(gaugeTimeout)) {
				if tmp, err := strconv.Atoi(n.Body); err != nil {
					fmt.Println("Error converting body to int: ", err)
				} else if tmp < 0 || tmp > 100 {
					fmt.Println("gauge not in acceptable range:", tmp)
				} else {
					event.Type = gauge
					event.Expires = n.Created.Add(gaugeTimeout)
					event.Sender = n.Sender
					event.Gauge = uint8(tmp)
					event.iconName = n.iconName
				}
			}
		} else if !n.isUrgent() {
			if event.Type != critical && now.Before(n.Created.Add(normalTimeout)) {
				if event.Type != normal {
					event.Type = normal
					event.Expires = n.Created.Add(normalTimeout)
					event.Sender = n.Sender
					event.Subject = n.Subject
					event.Body = []string{n.Body}
					event.Links = respond.Links{n.Link()}
					event.iconName = n.iconName
				} else if event.Sender == n.Sender && event.Subject == n.Subject && len(event.Body) < 3 {
					event.Body = append(event.Body, n.Body)
				}
			}
		} else {
			if event.Type != critical && now.Before(n.Created.Add(urgentTimeout)) {
				event.Type = critical
				event.Expires = n.Created.Add(urgentTimeout)
				event.Sender = n.Sender
				event.Subject = n.Subject
				event.Body = []string{n.Body}
				event.Links = respond.Links{n.Link()}
				event.iconName = n.iconName
			}

		}
	}
	return event
}

func sendEvent(n *Notification) {
	watch.SomethingChanged(n.self)
}
