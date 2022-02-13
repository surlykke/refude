// Copyright (c) Christian Surlykke
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
			if res := Windows.Get(path); res != nil {
				var win = *(res.(*Window))
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
				Windows.Put(&win)
				if relevantForDesktopSearch(&win) {
					watch.DesktopSearchMayHaveChanged()
				}
			}
		}
	}
}

var Windows = resource.MakeCollection()

func updateWindowList(p x11.Proxy) (somethingChanged bool) {
	var wIds = x11.GetStack(p)
	var oldResources = Windows.GetAll()
	var newResources = make([]resource.Resource, len(wIds), len(wIds))
	for i, wId := range wIds {
		var path = fmt.Sprintf("/window/%X", wId)
		var win *Window = nil
		for _, o := range oldResources {
			if path == o.Self() {
				win = o.(*Window)
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
		newResources[i] = win
	}
	if somethingChanged {
		Windows.ReplaceWith(newResources)
	}
	return
}

func showAndRaise(id uint32) {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	x11.MapAndRaiseWindow(requestProxy, id)
}

// --------------------------- Serving http requests -------------------------------

// http requests are concurrent, so all access to x11 from handling an http request, happens through
// this
var requestProxy = x11.MakeProxy()

// - and uses this for synchronization
var requestProxyMutex sync.Mutex
