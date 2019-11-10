// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"log"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"

	"github.com/surlykke/RefudeServices/lib/resource"
)

const (
	NET_WM_VISIBLE_NAME      = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME              = "_NET_WM_NAME"
	WM_NAME                  = "WM_NAME"
	NET_WM_ICON              = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE             = "_NET_WM_STATE"
)

// Maintains windows lists
func Run() {
	fmt.Println("Ind i window.Run")
	var eventConnection = xlib.MakeConnection()
	eventConnection.SubscribeToStackEvents()

	for {
		var wIds, err = eventConnection.GetUint32s(0, NET_CLIENT_LIST_STACKING)
		if err != nil {
			log.Println("WARN: Unable to retrieve _NET_CLIENT_LIST_STACKING", err)
			wIds = []uint32{}
		}
		mapWindowResources(wIds)
		eventConnection.WaitforStackEvent()
	}
}

func mapWindowResources(wIds []uint32) {
	// Reverse to get top window first
	for i, j := 0, len(wIds)-1; i < j; i, j = i+1, j-1 {
		wIds[i], wIds[j] = wIds[j], wIds[i]
	}
	var resources = make(map[string]interface{}, 2*len(wIds)+2)
	var windows = make([]*Window, 0, len(wIds))
	for i, wId := range wIds {
		var window = MakeWindow(wId)
		window.stackOrder = -i
		resources[window.Self] = window
		windows = append(windows, window)
		resources[ScreenshotSelf(wId)] = ScreenShot(wId)
	}
	resources["/windows"] = windows
	resource.MapCollection(&resources, "windows")
}
