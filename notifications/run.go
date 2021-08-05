// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"strconv"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/resource"
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

func GetResource(relPath []string) resource.Resource {
	if len(relPath) == 1 {
		if relPath[0] == "list" {
			var collection = resource.Collection{resource.MakeLink("/notification/list", "Notifications", "", relation.Self)}
			for _, n := range getNotifications() {
				collection = append(collection, resource.MakeLink(n.self, n.Subject, n.iconName, relation.Related))
			}
			return collection
		} else if relPath[0] == "flash" {
			if f := getFlash(); f != nil {
				return f
			}
		} else if id, err := strconv.Atoi(relPath[0]); err == nil {
			if n := getNotification(uint32(id)); n != nil {
				return n
			}
		}
	}
	return nil
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	var notifications = getNotifications()
	for _, notification := range notifications {
		if !forDisplay || !notification.forDisplay() {
			crawler(notification.self, notification.Subject, notification.iconName)
		}
	}
}
