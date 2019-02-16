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
	"strconv"
	"strings"
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
	mutex sync.Mutex
	notifications map[uint32]*Notification
	server.JsonResponseCache2
	server.PatchNotAllowed
	server.DeleteNotAllowed
}


func (*NotificationsCollection) HandledPrefixes() []string {
	return []string{"/notification"}
}

func MakeNotificationsCollection() *NotificationsCollection {
	var nc = &NotificationsCollection{}
	nc.JsonResponseCache2 = server.MakeJsonResponseCache2(nc)
	nc.notifications = make(map[uint32]*Notification)
	return nc
}

func (pc *NotificationsCollection) POST(w http.ResponseWriter, r *http.Request) {
	if res, err := pc.GetResource(r); err != nil {
		requests.ReportUnprocessableEntity(w, err)
	} else if res == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if notification, ok := res.(*Notification); !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func (dac *NotificationsCollection) GetResource(r *http.Request) (interface{}, error) {
	dac.mutex.Lock()
	defer dac.mutex.Unlock()

	var path = r.URL.Path
	if path == "/notifications" {
		var notifications = make([]*Notification, 0, len(dac.notifications))

		var matcher, err = requests.GetMatcher(r);
		if err != nil {
			return nil, err
		}

		for _, notification := range dac.notifications {
			if matcher(notification) {
				notifications = append(notifications, notification)
			}
		}

		return notifications, nil
	} else if strings.HasPrefix(path, "/notification/") {
		if id, err := strconv.ParseUint(path[len("/notification/"):], 10, 32); err != nil {
			return nil, nil
		} else if notification, ok := dac.notifications[uint32(id)]; ok {
			return notification, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}

}


