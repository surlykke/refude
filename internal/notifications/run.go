// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"fmt"

	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/repo"
	"github.com/surlykke/refude/internal/notifygui"
	"github.com/surlykke/refude/internal/watch"
)

var NotificationMap = repo.MakeSynkMap[uint32, *Notification]()

func removeNotification(id uint32, reason uint32) {
	if n, ok := NotificationMap.Get(id); ok && !n.Deleted {
		var copy = *n
		copy.Deleted = true
		NotificationMap.Put(id, &copy)
		conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
		watch.Publish("resourceChanged", "/flash")
		sendNotificationsToGui()
		watch.Publish("search", "")
	}
}

func getFlash() (map[string]string, bool) {
	var notifications = NotificationMap.GetAll()
	for i := len(notifications) - 1; i >= 0; i-- {
		n := notifications[i]
		if n.Deleted {
			continue
		}
		if !n.SoftExpired() {
			return map[string]string{
				"subject":      n.Title,
				"body":         n.Body,
				"iconFilePath": icons.FindIcon(string(n.iconName), uint32(64)),
			}, true
		}

	}
	return nil, false
}

func sendNotificationsToGui() {
	fmt.Println("Sending notifications to gui")
	var notificationsAsStrings = make([][]string, 0, 20)
	for _, n := range NotificationMap.GetAll() {
		if n.Deleted || n.SoftExpired() {
			continue
		}
		notificationsAsStrings = append(notificationsAsStrings, []string{n.Title, n.Body, icons.FindIcon(string(n.iconName), uint32(64))})
	}
	fmt.Println("Sending:", notificationsAsStrings)
	notifygui.SendNotificationsToGui(notificationsAsStrings)
}
