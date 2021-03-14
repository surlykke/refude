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

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/notifications/critical" {
		return respond.MakeRelatedCollection("/notifications", "Notifications", collectLinks(getCriticalNotifications()))
	} else if r.URL.Path == "/notifications" {
		return respond.MakeRelatedCollection("/notifications", "Notifications", collectLinks(getNotifications()))
	} else if r.URL.Path == flashPath {
		if n := getFlash(); n != nil {
			return n
		}
	} else if strings.HasPrefix(r.URL.Path, "/notification/") {
		if id, err := strconv.Atoi(r.URL.Path[len("/notification/"):]); err == nil {
			if notification := getNotification(uint32(id)); notification != nil {
				return notification
			}
		}
	}

	return nil
}

func collectLinks(list []*Notification) []respond.Link {
	var links = make([]respond.Link, 0, len(list))
	for _, n := range list {
		links = append(links, n.GetRelatedLink(0))
	}
	return links
}

func DesktopSearch(term string, baserank int) []respond.Link {
	var notifications = getNotifications()
	var links = make([]respond.Link, 0, len(notifications))
	for _, notification := range notifications {
		if len(notification.Self.Options.POST) > 0 {
			var rank int
			var ok bool
			if rank, ok = searchutils.Rank(notification.Subject, term, baserank); !ok {
				rank, ok = searchutils.Rank(notification.Body, term, baserank+100)
			}
			if ok {
				links = append(links, notification.GetRelatedLink(rank))
			}
		}
	}
	return links
}

func AllPaths() []string {
	var notifications = getNotifications()
	var paths = make([]string, 0, len(notifications)+2)
	for _, n := range notifications {
		paths = append(paths, n.Self.Href)
	}
	paths = append(paths, "/notifications")
	paths = append(paths, "/notification/osd")
	return paths
}
