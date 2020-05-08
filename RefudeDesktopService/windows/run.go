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
	"strings"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
	"github.com/surlykke/RefudeServices/lib/requests"
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
	if r.URL.Path == "/windows" {
		respond.AsJson(w, r, Collect(searchutils.Term(r)))
	} else if window, ok := getWindow(r.URL.Path); ok {
		if r.Method == "POST" {
			dataMutex.Lock()
			defer dataMutex.Unlock()
			dataConnection.RaiseAndFocusWindow(window.id)
			respond.Accepted(w)
		} else {
			respond.AsJson(w, r, window.ToStandardFormat())
		}
	} else if window, ok := getWindowForScreenShot(r.URL.Path); ok {
		if r.Method == "GET" {
			var downscaleS = requests.GetSingleQueryParameter(r, "downscale", "1")
			var downscale = downscaleS[0] - '0'
			if downscale < 1 || downscale > 5 {
				respond.UnprocessableEntity(w, fmt.Errorf("downscale should be >= 1 and <= 5"))
			} else if bytes, err := getScreenshot(window.id, downscale); err == nil {
				w.Header().Set("Content-Type", "image/png")
				w.Write(bytes)
			} else {
				respond.ServerError(w, err)
			}
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func getWindowForScreenShot(path string) (Window, bool) {
	if strings.HasSuffix(path, "/screenshot") {
		return getWindow(path[0 : len(path)-11])
	}
	return Window{}, false
}

func getWindow(path string) (Window, bool) {
	for _, w := range windows.Load().([]Window) {
		if w.self == path {
			return w, true
		}
	}
	return Window{}, false
}

func Collect(term string) respond.StandardFormatList {
	var winList = windows.Load().([]Window)
	var sfl = make(respond.StandardFormatList, 0, len(winList))
	for _, win := range winList {
		var sf = win.ToStandardFormat()
		if rank := searchutils.SimpleRank(sf.Title, "", term); rank > -1 {
			sfl = append(sfl, sf)
		}
	}

	return sfl
}

func SearchWindows(collector *searchutils.Collector) {
	for _, wi := range windows.Load().([]Window) {
		collector.Collect(wi.ToStandardFormat())
	}
}

func AllPaths() []string {
	var windowList = windows.Load().([]Window)
	var paths = make([]string, 0, 2*len(windowList)+1)
	for _, window := range windowList {
		paths = append(paths, window.self)
		paths = append(paths, window.self+"/screenshot")
	}
	paths = append(paths, "/windows")
	return paths
}

// Maintains windows lists
func Run() {
	var eventConnection = xlib.MakeConnection()
	eventConnection.SubscribeToStackEvents()

	for {
		if wIds, err := eventConnection.GetUint32s(0, NET_CLIENT_LIST_STACKING); err != nil {
			log.Println("WARN: Unable to retrieve _NET_CLIENT_LIST_STACKING", err)
			windows.Store([]Window{})
		} else {
			var list = make([]Window, len(wIds), len(wIds))
			// Revert so highest in stach comes first
			for i := 0; i < len(wIds); i++ {
				list[i] = Window{self: fmt.Sprintf("/window/%d", wIds[len(wIds)-i-1]), id: wIds[len(wIds)-i-1]}
			}
			windows.Store(list)
		}

		eventConnection.WaitforStackEvent()
	}
}

var windows atomic.Value

func init() {
	windows.Store([]Window{})
}
