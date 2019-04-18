// Copyright (c) 2017 Christian Surlykke
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

	"github.com/surlykke/RefudeServices/lib/resource"
)

const NotificationMediaType resource.MediaType = "application/vnd.org.refude.desktopnotification+json"

var notifications = make(map[resource.StandardizedPath]*Notification)
var lock sync.Mutex

func GetNotification(path resource.StandardizedPath) *Notification {
	lock.Lock()
	defer lock.Unlock()

	return notifications[path]
}

func GetNotifications() []resource.Resource {
	lock.Lock()
	defer lock.Unlock()

	var result = make([]resource.Resource, len(notifications), len(notifications))
	var i = 0
	for _, notification := range notifications {
		result[i] = notification
		i++
	}
	return result
}

func setNotification(notification *Notification) {
	lock.Lock()
	defer lock.Unlock()

	notifications[notification.GetSelf()] = notification
}

func removeNotification(path resource.StandardizedPath, internalId uint32) bool {
	lock.Lock()
	defer lock.Unlock()

	if notification := notifications[path]; notification != nil && (internalId == 0 || internalId == notification.internalId) {
		delete(notifications, path)
		return true
	} else {
		return false
	}
}

type Notification struct {
	resource.GenericResource
	Id            uint32
	internalId    uint32
	Sender        string
	Subject       string
	Body          string
	RelevanceHint int
	Expires       *time.Time `json:",omitempty"`
}

func (n *Notification) removeAfter(duration time.Duration) {
	time.AfterFunc(duration, func() { removals <- removal{n.Id, n.internalId, Expired} })
}

func (nc *Notification) DELETE(w http.ResponseWriter, r *http.Request) {
	// FIXME
}

func notificationSelf(id uint32) resource.StandardizedPath {
	return resource.Standardizef("/notification/%d", id)
}
