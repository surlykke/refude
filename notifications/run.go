// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"strconv"

	"github.com/surlykke/RefudeServices/lib/link"
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

func GetResource(relPath string) resource.Resource {
	if relPath == "list" {
		var ll = link.MakeList("/notification/list", "Notifications", "")
		for _, n := range getNotifications() {
			ll = ll.Add(n.self, n.Subject, n.iconName, relation.Related)
		}
		return link.Collection(ll)
	} else if relPath == "flash" {
		if f := getFlash(); f != nil {
			return f
		}
	} else if id, err := strconv.Atoi(relPath); err == nil {
		if n := getNotification(uint32(id)); n != nil {
			return n
		}
	}
	return nil
}

func Collect(term string, sink chan link.Link) {
	for _, n := range notifications {
		if n.forDisplay() {
			if rnk := searchutils.Match(term, n.Subject); rnk > -1 {
				sink <- link.MakeRanked(n.self, n.Subject, n.iconName, "notification", rnk)
			}
		}
	}
}

func CollectPaths(method string, sink chan string) {
	sink <- "/notification/list"
	sink <- "/notification/flash"

	for _, notification := range getNotifications() {
		sink <- notification.self
	}
}
