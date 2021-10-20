// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/windows/x11"
)

// Maintains windows  and monitors lists
func Run() {
	var proxy = x11.MakeProxy()
	x11.SubscribeToEvents(proxy)

	updateDesktopLayout(proxy)
	updateWindowList(proxy)

	for {
		event, wId := x11.NextEvent(proxy)
		if event == x11.DesktopGeometry {
			updateDesktopLayout(proxy)
		} else if event == x11.DesktopStacking {
			if updateWindowList(proxy) {
				watch.DesktopSearchMayHaveChanged()
			}
		} else {
			// So it's a 'single'-window event
			var path string = fmt.Sprintf("/window/%X", wId)
			if data := Windows.GetData(path); data != nil {
				var win = data.(*Window).shallowCopy()
				switch event {
				case x11.WindowTitle:
					win.Name, _ = x11.GetName(proxy, wId)
				case x11.WindowIconName:
					win.IconName, _ = GetIconName(proxy, wId)
				case x11.WindowSt:
					win.State = x11.GetStates(proxy, wId)
				// TODO case x11.WindowGeometry:
				default:
					continue
				}
				Windows.MakeAndPut(path, win.Name, "", win.IconName, win)
				if relevantForDesktopSearch(win) {
					watch.DesktopSearchMayHaveChanged()
				}
			}
		}
	}
}

var Windows = resource.MakeList("window", true, "/window/list", 20)

func updateWindowList(p x11.Proxy) (somethingChanged bool) {
	var wIds = x11.GetStack(p)
	var oldResources = Windows.GetAll()
	var newResources = make([]resource.Resource, len(wIds), len(wIds))
	for i, wId := range wIds {
		var path = fmt.Sprintf("/window/%X", wId)
		var win *Window = nil
		for _, o := range oldResources {
			if path == o.Path {
				win = o.Data.(*Window).shallowCopy()
				break
			}
		}
		if win == nil {
			win = makeWindow(p, wId)
			x11.SubscribeToWindowEvents(p, wId)
		}
		if win.Stacking != i {
			win.Stacking = i
			somethingChanged = true
		}
		newResources[i] = resource.MakeResource(path, win.Name, "", win.IconName, "window", win)
	}
	if somethingChanged {
		Windows.ReplaceWith(newResources)
	}
	return
}

func clientWindowIds() []uint32 {
	var result []uint32
	for _, res := range Windows.GetAll() {
		if res.Data.(*Window).Name == "org.refude.client" {
			result = append(result, res.Data.(*Window).Id)
		}
	}
	return result
}

func ShowClientWindow() bool {
	var found bool = false
	for i, id := range clientWindowIds() {
		found = true
		if i == 0 {
			showAndRaise(id)
		} else {
			performDelete(id)
		}
	}

	return found
}

func showAndRaise(id uint32) {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	x11.MapAndRaiseWindow(requestProxy, id)
}

func CloseClientWindow() {
	for _, id := range clientWindowIds() {
		performDelete(id)
	}
}

// --------------------------- Serving http requests -------------------------------

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var requestProxy = x11.MakeProxy()

// - and uses this for synchronization
var requestProxyMutex sync.Mutex
