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

func StartGui() {
	go C.run()
}

var notificationsChan = make(chan [][]string, 20)

func SendNotificationsToGui(notifications [][]string) {
	notificationsChan <- notifications

}

//export GuiReady
func GuiReady() {
	go sendNotificationsToGui()
}

func sendNotificationsToGui() {
	for notifications := range notificationsChan {
		var cStrings = make([]*C.char, 0, 100)
		for _, n := range notifications {
			cStrings = append(cStrings, C.CString(n[0]), C.CString(n[1]), C.CString(n[2]))
		}
		if len(cStrings) > 0 {
			C.update(&cStrings[0], C.int(len(cStrings)/3))
		} else {
			C.update(nil, C.int(0))
		}
	}
}
