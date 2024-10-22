// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/watch"
)

func removeNotification(id uint32, reason uint32) {
	if n, ok := repo.Get[*Notification](path.Of("/notification/%d", id)); ok && !n.Deleted {
		var copy = *n
		copy.Deleted = true
		repo.Put(&copy)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		watch.Publish("resourceChanged", "/flash")
		watch.Publish("search", "")
	}
}

func getFlash() (map[string]string, bool) {
	var notifications = repo.GetListSortedByPath[*Notification]("/notification/")
	for i := len(notifications) - 1; i >= 0; i-- {
		n := notifications[i]
		if n.Deleted {
			continue
		}
		if !n.SoftExpired() {
			return map[string]string{
				"subject":      n.Title,
				"body":         n.Comment,
				"iconFilePath": icons.FindIcon(string(n.iconName), uint32(64)),
			}, true
		}
	}
	return nil, false
}
