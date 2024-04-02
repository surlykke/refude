// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package x11

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type X11Window struct {
	resource.BaseResource
	Wid              uint32
	State            WindowStateMask
	ApplicationName  string `json:"applicationName"`
	ApplicationClass string `json:"applicationClass"`
}

func MakeWindow(p Proxy, wId uint32) (*X11Window, error) {
	if name, err := GetName(p, wId); err != nil {
		return nil, err
	} else {
		var iconName, _ = getIconName(p, wId)
		var state, _ = GetStates(p, wId)
		var appName, appClass = GetApplicationAndClass(p, wId)
		return &X11Window{
			BaseResource: resource.BaseResource{
				Path:     fmt.Sprintf("/window/%d", wId),
				Title:    name,
				Comment:  appClass,
				IconUrl: link.IconUrl(iconName),
				Profile:  "window",
			},
			Wid:              wId,
			State:            state,
			ApplicationName:  appName,
			ApplicationClass: appClass,
		}, nil
	}

}

func (this *X11Window) Actions() link.ActionList {
	var iconName = this.IconUrl
	return link.ActionList{{Name: "activate", Title: "Raise and focus", IconUrl: iconName}}
}

func (this *X11Window) DeleteAction() (string, bool) {
	return "Close", true
}

func (this *X11Window) RelevantForSearch(term string) bool {
	return !strings.HasPrefix(this.Title, "Refude launcher") &&
		this.State&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) == 0
}

func (this X11Window) DoDelete(w http.ResponseWriter, r *http.Request) {
	this.Close()
	respond.Accepted(w)
}

func (this X11Window) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "activate")
	if "activate" == action {
		this.RaiseAndFocus()
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

func (this X11Window) RaiseAndFocus() {
	proxy.Lock()
	defer proxy.Unlock()

	RaiseAndFocusWindow(proxy, uint32(this.Wid))
}

func (this X11Window) Close() {
	proxy.Lock()
	defer proxy.Unlock()

	CloseWindow(proxy, uint32(this.Wid))
}

func (this X11Window) GetIconName(p Proxy) string {
	p.Lock()
	defer p.Unlock()

	if name, err := GetIconName(proxy, uint32(this.Wid)); err != nil {
		return ""
	} else {
		return name
	}
}

func GetIconName(p Proxy, wId uint32) (string, error) {
	pixelArray, err := GetIcon(p, wId)
	if err != nil {
		log.Warn("Error retrieving x11 icon", err)
		return "", err
	} else {
		return icons.AddX11Icon(pixelArray)
	}
}

var proxy = MakeProxy()

func Run() {
	var proxy = MakeProxy()
	SubscribeToEvents(proxy)
	updateWindowList(proxy)

	for {
		event, wId := NextEvent(proxy)
		if event == DesktopStacking {
			updateWindowList(proxy)
		} else if event == WindowTitle {
			updateWindow(proxy, wId, titleUpdater)
		} else if event == WindowIconName {
			updateWindow(proxy, wId, iconUpdater)
		} else if event == WindowSt {
			updateWindow(proxy, wId, stateUpdater)
		} else {
			continue
		}
	}
}

func updateWindowList(p Proxy) {
	var wIds = GetStack(p)
	var xWins = make([]*X11Window, 0, len(wIds))
	for _, wId := range wIds {
		if x11Window, err := MakeWindow(p, wId); err == nil {
			xWins = append(xWins, x11Window)
			SubscribeToWindowEvents(p, wId)
		}
	}
	resourcerepo.ReplacePrefixWithList("/window/", xWins)
}

func updateWindow(p Proxy, wId uint32, updater func(Proxy, *X11Window) bool) {
	if w, ok := resourcerepo.GetTyped[*X11Window](fmt.Sprintf("/window/%d", wId)); ok {
		var copy = *w
		if updater(p, &copy) {
			resourcerepo.Update(&copy)
		}
	}
}

func titleUpdater(p Proxy, win *X11Window) bool {
	if title, err := GetName(p, win.Wid); err == nil {
		win.Title = title
		return true
	} else {
		return false
	}
}

func iconUpdater(p Proxy, win *X11Window) bool {
	if iconName, err := getIconName(p, win.Wid); err == nil {
		win.IconUrl = link.IconUrl(iconName)
		return true
	} else {
		return false
	}
}

func stateUpdater(p Proxy, win *X11Window) bool {
	if state, err := GetStates(p, win.Wid); err == nil {
		win.State = state
		return true
	} else {
		return false
	}
}

func getIconName(p Proxy, wId uint32) (string, error) {
	if pixelArray, err := GetIcon(p, uint32(wId)); err != nil {
		return "", err
	} else if name, err := icons.AddX11Icon(pixelArray); err != nil {
		return "", err
	} else {
		return name, nil
	}
}

func GetMonitors() []*MonitorData {
	proxy.Lock()
	defer proxy.Unlock()
	return GetMonitorDataList(proxy)
}
