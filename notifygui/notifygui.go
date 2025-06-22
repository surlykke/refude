package notifygui

/*
  #cgo pkg-config: gtk4 gtk4-layer-shell-0
 #include <stdio.h>
 #include <stdlib.h>
 #include "notifygui.h"
*/
import "C"
import "fmt"

func StartGui() {
	fmt.Println("StartGui")
	go C.run()
}

func SendNotificationsToGui(notifications [][]string) {
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
