// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/watch"
)

var notificationRepo = repo.MakeRepoWithFilter[*Notification](searchFilter)

var added = make(chan *Notification)
var removals = make(chan notificationRemoval)

func Run() {
	go DoDBus()
	var Requests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case req := <-Requests:
			if fn := getFlash(req); fn != nil {
				req.Replies <- resource.RankedResource{Rank: 0, Res: fn}
			}
			notificationRepo.DoRequest(req)
		case n := <-added:
			notificationRepo.Put(n)
			watch.Publish("resourceChanged", "/flash")
			if n.Urgency == Low {
				time.AfterFunc(2050*time.Millisecond, func() { watch.Publish("resourceChanged", "/flash") })
			} else if n.Urgency == Normal {
				time.AfterFunc(10050*time.Millisecond, func() { watch.Publish("resourceChanged", "/flash") })
			}
		case removal := <-removals:
			removeNotification(removal)
			watch.Publish("resourceChanged", "/flash")
			watch.Publish("/notification/", "")
		}
	}
}

type notificationRemoval struct {
	id     uint32
	reason uint32
}

func removeNotification(removal notificationRemoval) {
	if n, ok := notificationRepo.Get(fmt.Sprintf("/notification/%d", removal.id)); ok && !n.Deleted {
		var copy = *n
		copy.Deleted = true
		notificationRepo.Put(&copy)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", removal.id, removal.reason)
	}
}

func searchFilter(term string, n *Notification) bool {
	if len(term) >= 3 {
		return true
	} else if _, ok := n.NActions["default"]; ok && n.NotExpired() {
		return true
	} else {
		return false
	}
}

type Flash struct {
	resource.ResourceData
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	IconFilePath string `json:"iconFilePath"`
	Urgency      Urgency
}

func getFlash(req repo.ResourceRequest) *Flash {
	var fn *Notification = nil
	if req.ReqType == repo.ByPath && req.Data == "/flash" || req.ReqType == repo.ByPathPrefix && strings.HasPrefix("/flash", req.Data) {

		var now = time.Now()
		var notifications = notificationRepo.GetAll()

		for i := len(notifications) - 1; i >= 0; i-- {
			var n = notifications[i]
			if n.Deleted {
				continue
			}

			if n.Urgency == Critical {
				fn = n
				break
			} else if n.Urgency == Normal {
				if fn == nil || fn.Urgency < Normal {
					if now.Before(time.Time(n.Created).Add(6 * time.Second)) {
						fn = n
					}
				}
			} else { /* n.Urgency == Low */
				if fn == nil && now.Before(time.Time(n.Created).Add(2*time.Second)) {
					fn = n
				}
			}
		}
	}
	var flash *Flash = nil
	if fn != nil {
		flash = &Flash{Subject: fn.Title, Body: fn.Comment, IconFilePath: icons.FindIconPath(fn.iconName, 64), Urgency: fn.Urgency}
	}

	return flash
}
