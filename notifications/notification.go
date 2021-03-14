// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

type Urgency string

const (
	critical Urgency = "Critical"
	normal           = "Normal"
	low              = "Low"
)

type Notification struct {
	respond.Resource
	Id       uint32
	Sender   string
	Subject  string
	Body     string
	Created  time.Time
	Expires  time.Time `json:",omitempty"`
	Urgency  Urgency
	Actions  map[string]string
	Hints    map[string]interface{}
	iconName string
}

func (n *Notification) haveDefaultAction() bool {
	_, ok := n.Actions["default"]
	return ok
}

func (n *Notification) copy() *Notification {
	var res = &Notification{}
	*res = *n
	res.Owner = res
	return res
}

type SortableNotificationList []*Notification

func (snl SortableNotificationList) Len() int {
	return len(snl)
}

func (snl SortableNotificationList) Less(i int, j int) bool {
	return snl[i].Created.After(snl[j].Created) // Latest first
}

func (snl SortableNotificationList) Swap(i int, j int) {
	snl[i], snl[j] = snl[j], snl[i]
}

const flashPath = "/notification/flash"

type Flash struct {
	respond.Resource
	NotificationId uint32
	Sender         string
	Subject        string
	Body           []string
	Created        time.Time
	Expires        time.Time
}

func MakeFlash(n *Notification) *Flash {
	var f = &Flash{
		NotificationId: n.Id,
		Sender:         n.Sender,
		Subject:        n.Subject,
		Body:           []string{n.Body},
		Created:        n.Created,
		Expires:        n.Created.Add(time.Second * 6),
	}
	var icon = n.Self.Icon
	if icon == "" {
		icon = icons.IconUrl("dialog-info")
	}
	f.Resource = respond.MakeResource(flashPath, n.Subject, icon, f)
	f.AddLink(n.GetRelatedLink(0))
	return f
}

func (f *Flash) differentFrom(other *Flash) bool {
	if f == nil {
		return other != nil
	} else {
		return other == nil || f.NotificationId != other.NotificationId
	}
}

// Notifiation collection

var (
	lock          sync.Mutex
	notifications = make(map[uint32]*Notification)
	flash         *Flash
)

func getCriticalNotifications() []*Notification {
	lock.Lock()
	defer lock.Unlock()
	return extractNotifications(2)
}

func getNotifications() []*Notification {
	lock.Lock()
	defer lock.Unlock()
	return extractNotifications(0)
}

func getNotification(id uint32) *Notification {
	lock.Lock()
	defer lock.Unlock()
	return notifications[id]
}

func setNotification(n *Notification) {
	lock.Lock()
	defer lock.Unlock()
	fmt.Println("Setting notification:", string(respond.ToJson(n)))
	if n.Urgency != critical {
		// Only flash for nonCriticalNotifications, clients should look at /notifications/critical to see critical notifications
		var newFlash = MakeFlash(n)
		for _, n := range extractNotifications(1) {
			if n.Sender == newFlash.Sender && n.Subject == newFlash.Subject {
				newFlash.Body = append(newFlash.Body, n.Body)
			}
		}
		_doJustAfter(n.Expires, removeExpired)
		if flash == nil {
			_doJustAfter(newFlash.Expires, removeFlash)
		}
		flash = newFlash
	}
	notifications[n.Id] = n

	watch.DesktopSearchMayHaveChanged()
	watch.SomethingChanged(flashPath)
}

func removeNotification(id uint32, reason uint32) {
	lock.Lock()
	defer lock.Unlock()
	var ok = false
	if _, ok = notifications[id]; ok {
		delete(notifications, id)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		watch.DesktopSearchMayHaveChanged()
	}

}

func removeExpired() {
	lock.Lock()
	defer lock.Unlock()
	var haveRemovals bool
	var now = time.Now()
	for id, n := range notifications {
		if n.Expires.Before(now) {
			delete(notifications, id)
			haveRemovals = true
			conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", n.Id, Expired)
		}
	}
	if haveRemovals {
		watch.DesktopSearchMayHaveChanged()
	}
}

func removeFlash() {
	lock.Lock()
	defer lock.Unlock()
	if flash.Expires.Before(time.Now()) {
		flash = nil
		watch.SomethingChanged(flashPath)
	} else {
		_doJustAfter(flash.Expires, removeFlash)
	}
}

func getFlash() *Flash {
	lock.Lock()
	defer lock.Unlock()
	return flash
}

func _doJustAfter(t time.Time, f func()) {
	time.AfterFunc(t.Sub(time.Now())+20*time.Millisecond, f)
}

// ------ Don't call directly, caller must have lock.

func extractNotifications(filter int /* 0: all, 1: non-critical, 2: critical*/) []*Notification {
	var list = make([]*Notification, 0, 2)
	for _, n := range notifications {
		if filter == 0 || filter == 1 && n.Urgency != critical || filter == 2 && n.Urgency == critical {
			list = append(list, n)
		}
	}
	sort.Sort(SortableNotificationList(list))
	return list
}
