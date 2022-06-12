// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"encoding/json"
	"net/http"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type XWin uint32

func (w XWin) Id() uint32 {
	return uint32(w)
}

func (w XWin) Presentation() (title string, comment string, icon link.Href, profile string) {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()
	var name, _ = x11.GetName(synchronizedProxy, uint32(w))
	var iconName, _ = GetIconName(synchronizedProxy, uint32(w))
	return name, "", link.IconUrl(iconName), "window"
}

func (w XWin) Links(self, searchTerm string) link.List {
	var name, iconName string
	var err error
	proxyMutex.Lock()
	name, err = x11.GetName(synchronizedProxy, uint32(w))
	iconName, _ = GetIconName(synchronizedProxy, uint32(w))
	proxyMutex.Unlock()

	if err == nil {
		var links = make(link.List, 0, 10)
		if rnk := searchutils.Match(searchTerm, name); rnk > -1 {
			links = append(links, link.Make(self, "Raise and focus", iconName, relation.DefaultAction))
			links = append(links, link.Make(self, "Close", iconName, relation.Delete))
		}
		var panes = collectPanes()
		for _, pane := range panes {
			if pane.XWinId == w.Id() {
				var title = pane.CurrentCommand + ":" + pane.CurrentDirectory
				if rnk := searchutils.Match(searchTerm, title, "pane", "tmux"); rnk > -1 {
					links = append(links, link.MakeRanked("/tmux/"+pane.PaneId, title, "", "tmux", rnk))
				}
			}
		}
		return links
	}

	return link.List{}
}

func (w XWin) MarshalJSON() ([]byte, error) {
	proxyMutex.Lock()
	var data = fetchWindowData(synchronizedProxy, uint32(w))
	proxyMutex.Unlock()
	return json.Marshal(data)
}

var Windows = resource.MakeCollection[uint32, XWin]("/window/")

type windowData struct {
	WindowId uint32
	Name     string
	IconName string `json:",omitempty"`
	State    x11.WindowStateMask
	Stacking int // 0 means: on top, then 1, then 2 etc. -1 means we don't know (yet)
}

// Caller ensures thread safety (calls to x11)
func fetchWindowData(p x11.Proxy, wId uint32) *windowData {
	var win = &windowData{WindowId: wId}
	win.Name, _ = x11.GetName(p, wId)
	win.IconName, _ = GetIconName(p, wId)
	win.State = x11.GetStates(p, wId)
	win.Stacking = -1
	return win
}

func (xWin XWin) DoDelete(w http.ResponseWriter, r *http.Request) {
	proxyMutex.Lock()
	x11.CloseWindow(synchronizedProxy, uint32(xWin))
	proxyMutex.Unlock()
	respond.Accepted(w)
}

func (xWin XWin) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if performAction(uint32(xWin), action) {
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

func performAction(wId uint32, action string) bool {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	var found = true
	if action == "" {
		x11.RaiseAndFocusWindow(synchronizedProxy, wId)
	}
	return found
}

func RaiseAndFocusNamedWindow(name string) bool {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	if wId, found := findNamedWindow(synchronizedProxy, name); found {
		x11.RaiseAndFocusWindow(synchronizedProxy, wId)
		return true
	} else {
		return false
	}
}

func ResizeNamedWindow(name string, newWidth, newHeight uint32) bool {
	proxyMutex.Lock()
	defer proxyMutex.Unlock()

	if wId, found := findNamedWindow(synchronizedProxy, name); found {
		x11.Resize(synchronizedProxy, wId, newWidth, newHeight)
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
