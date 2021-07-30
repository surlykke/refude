// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var incomingNotifications = make(chan *Notification)
var removals = make(chan removal)
var cleaningHints = make(chan struct{})

func Run() {
	go DoDBus()

	for {
		select {
		case notification := <-incomingNotifications:
			setNotification(notification)
		case rem := <-removals:
			removeNotification(rem.id, rem.reason)
		case <-cleaningHints:
			removeExpired()
		}
	}
}

var notificationPathPattern = regexp.MustCompile("^/notification/(\\d+)$")

func GetJsonResource(r *http.Request) respond.JsonResource {
	if r.URL.Path == "/notifications/critical" {
		return nil // FIXME
	} else if r.URL.Path == flashPath {
		if f := getFlash(); f != nil {
			return f
		}
	} else if r.URL.Path == "/notifications" {
		var res = respond.MakeResource("/notifications", "notifications", "", "collection")
		lock.Lock()
		for _, notification := range notifications {
			res.Links = append(res.Links, notification.GetRelatedLink())
		}
		lock.Unlock()
		return &res
	} else if strings.HasPrefix(r.URL.Path, "/notification/") {
		if id, err := strconv.Atoi(r.URL.Path[len("/notification/"):]); err == nil {
			if notification := getNotification(uint32(id)); notification != nil {
				return notification
			}
		}
	}
	return nil
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	var notifications = getNotifications()
	for _, notification := range notifications {
		if !forDisplay || !notification.forDisplay() {
			crawler(&notification.Resource, nil)
		}
	}
}
