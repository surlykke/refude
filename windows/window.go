// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"net/http"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type Bounds struct {
	X, Y int32
	W, H uint32
}

type Window struct {
	WindowId uint32
	Name     string
	IconName string `json:",omitempty"`
	State    x11.WindowStateMask
	Stacking int // 0 means: on top, then 1, then 2 etc. -1 means we don't know (yet)
}

func (w *Window) Id() uint32 {
	return w.WindowId
}

func (w *Window) Presentation() (title string, comment string, icon link.Href, profile string) {
	return w.Name, "", link.IconUrl(w.IconName), "window"
}

func (w *Window) Links(self, searchTerm string) link.List {
	if searchutils.Match(searchTerm, w.Name) > -1 {
		return link.List{
			link.Make(self, "Raise and focus", w.IconName, relation.DefaultAction),
			link.Make(self, "Close", w.IconName, relation.Delete),
		}
	} else {
		return link.List{}
	}
}

// Caller ensures thread safety (calls to x11)
func makeWindow(p x11.Proxy, wId uint32) *Window {
	var win = &Window{WindowId: wId}
	win.Name, _ = x11.GetName(p, wId)
	win.IconName, _ = GetIconName(p, wId)
	win.State = x11.GetStates(p, wId)
	win.Stacking = -1
	return win
}

func (win *Window) DoDelete(w http.ResponseWriter, r *http.Request) {
	requestProxyMutex.Lock()
	x11.CloseWindow(requestProxy, win.WindowId)
	requestProxyMutex.Unlock()
	respond.Accepted(w)
}

func (win *Window) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if performAction(win.WindowId, action) {
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

func performAction(wId uint32, action string) bool {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()

	var found = true
	if action == "" {
		x11.RaiseAndFocusWindow(requestProxy, wId)
	}
	return found
}

func RaiseAndFocusNamedWindow(name string) bool {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()

	if wId, found := findNamedWindow(requestProxy, name); found {
		x11.RaiseAndFocusWindow(requestProxy, wId)
		return true
	} else {
		return false
	}
}

func ResizeNamedWindow(name string, newWidth, newHeight uint32) bool {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()

	if wId, found := findNamedWindow(requestProxy, name); found {
		x11.Resize(requestProxy, wId, newWidth, newHeight)
		return true
	} else {
		return false
	}
}

func findNamedWindow(proxy x11.Proxy, name string) (uint32, bool) {
	for _, wId := range x11.GetStack(proxy) {
		if windowName, err := x11.GetName(proxy, wId); err == nil && windowName == name {
			return wId, true
		}
	}
	return 0, false
}

func GetIconName(p x11.Proxy, wId uint32) (string, error) {
	pixelArray, err := x11.GetIcon(p, wId)
	if err != nil {
		log.Warn("Error converting x11 icon to pngs", err)
		return "", err
	} else {
		return icons.AddX11Icon(pixelArray)
	}
}
