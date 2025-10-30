// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifygui

/*
 #cgo pkg-config: gtk4 gtk4-layer-shell-0
 #include <stdio.h>
 #include <stdlib.h>
 #include "notifygui.h"
*/
import "C"
import (
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/notifications"
)

func StartGui() {
	go C.run()
}

//export GuiReady
func GuiReady() {
	go run()
}

func run() {
	var notificationEvents = notifications.NotificationMap.Events.Subscribe()
	sendNotificationsToGui()
	for {
		notificationEvents.Next()
		sendNotificationsToGui()
	}
}

func sendNotificationsToGui() {
	var notificationsAsStrings = make([][]string, 0, 20)
	for _, n := range notifications.NotificationMap.GetAll() {
		if n.Deleted || n.SoftExpired() {
			continue
		}
		notificationsAsStrings = append(notificationsAsStrings, []string{n.Title, n.Body, icons.FindIcon(n.IconName, uint32(64))})
	}

	var cStrings = make([]*C.char, 0, 100)
	for _, n := range notificationsAsStrings {
		cStrings = append(cStrings, C.CString(n[0]), C.CString(n[1]), C.CString(n[2]))
	}
	if len(cStrings) > 0 {
		C.update(&cStrings[0], C.int(len(cStrings)/3))
	} else {
		C.update(nil, C.int(0))
	}
}
