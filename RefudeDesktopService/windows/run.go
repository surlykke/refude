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
	"net/http"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

const (
	NET_WM_VISIBLE_NAME      = "_NET_WM_VISIBLE_NAME"
	NET_WM_NAME              = "_NET_WM_NAME"
	WM_NAME                  = "WM_NAME"
	NET_WM_ICON              = "_NET_WM_ICON"
	NET_CLIENT_LIST_STACKING = "_NET_CLIENT_LIST_STACKING"
	NET_WM_STATE             = "_NET_WM_STATE"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if wi, ok := windows.Load().(WindowMap)[r.URL.Path]; ok {
		if r.Method == "GET" {
			respond.AsJson(w, wi.ToStandardFormat())
		} else if r.Method == "POST" {
			dataMutex.Lock()
			defer dataMutex.Unlock()
			dataConnection.RaiseAndFocusWindow(uint32(wi))
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func SearchWindows(collector *searchutils.Collector) {
	for _, wi := range windows.Load().(WindowMap) {
		collector.Collect(wi.ToStandardFormat())
	}
}

func AllPaths() []string {
	var vm = windows.Load().(WindowMap)
	var paths = make([]string, 0, len(vm))
	for path, _ := range vm {
		paths = append(paths, path)
	}
	return paths
}

// Maintains windows lists
func Run() {
	fmt.Println("Ind i window.Run")
	var eventConnection = xlib.MakeConnection()
	eventConnection.SubscribeToStackEvents()

	for {
		wIds, err := eventConnection.GetUint32s(0, NET_CLIENT_LIST_STACKING)
		if err != nil {
			log.Println("WARN: Unable to retrieve _NET_CLIENT_LIST_STACKING", err)
			wIds = []uint32{}
		}

		var wm = make(WindowMap, len(wIds))
		for _, wId := range wIds {
			wm[fmt.Sprintf("/window/%d", wId)] = Window(wId)
		}
		windows.Store(wm)

		eventConnection.WaitforStackEvent()
	}
}

var windows atomic.Value

func init() {
	windows.Store(make(WindowMap))
}
