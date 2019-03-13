// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"net/http"
	"sync"
	"time"
)

const NotificationMediaType resource.MediaType = "application/vnd.org.refude.desktopnotification+json"

type Notification struct {
	resource.AbstractResource
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

type NotificationsCollection struct {
	mutex         sync.Mutex
	notifications map[resource.StandardizedPath]*Notification
	server.CachingJsonGetter
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func (*NotificationsCollection) HandledPrefixes() []string {
	return []string{"/notification"}
}

func MakeNotificationsCollection() *NotificationsCollection {
	var nc = &NotificationsCollection{}
	nc.CachingJsonGetter = server.MakeCachingJsonGetter(nc)
	nc.notifications = make(map[resource.StandardizedPath]*Notification)
	return nc
}

func (nc *NotificationsCollection) POST(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/notifications" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if res := nc.GetSingle(r); res == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if notification, ok := res.(*Notification); !ok {
		w.WriteHeader(http.StatusMethodNotAllowed) // Shouldn't happen
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "")
		if action, ok := notification.ResourceActions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}

func (nc *NotificationsCollection) GetSingle(r *http.Request) interface{} {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()
	if n, ok := nc.notifications[resource.Standardize(r.URL.Path)]; ok {
		return n
	} else {
		return nil
	}
}

func (nc *NotificationsCollection) GetCollection(r *http.Request) []interface{} {
	nc.mutex.Lock()
	defer nc.mutex.Unlock()

	if r.URL.Path == "/notifications" {
		var result = make([]interface{}, 0, len(nc.notifications))
		for _, app := range nc.notifications {
			result = append(result, app)
		}
		return result
	} else {
		return nil
	}
}

func notificationSelf(id uint32) resource.StandardizedPath {
	return resource.Standardizef("/notification/%d", id)
}
