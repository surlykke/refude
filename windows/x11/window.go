// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package x11

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/windows/monitor"
)

type X11Window uint32

func (this X11Window) Id() uint32 {
	return uint32(this)
}

func (this X11Window) GetPath() string {
	return strconv.Itoa(int(this))
}

func (this X11Window) Presentation() (title string, comment string, iconName string, profile string) {
	return this.GetName(), "", this.GetIconName(), "window"
}

func (this X11Window) Actions() link.ActionList {
	var iconName = this.GetIconName()
	return link.ActionList{{Name: "activate", Title: "Raise and focus", IconName: iconName}}
}

func (this X11Window) DeleteAction() (string, bool) {
	return "Close", true
}

func (this X11Window) Links(searchTerm string) link.List {
	return link.List{}
}

func (this X11Window) RelevantForSearch() bool {
	var applicationName, _ = this.GetApplicationAndClass()
	var states = this.GetStates()
	return applicationName != "localhost__refude_html_launcher" &&
		applicationName != "localhost__refude_html_notifier" &&
		states&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) == 0
}

type windowData struct {
	WindowId         uint32
	Name             string
	IconName         string `json:",omitempty"`
	State            WindowStateMask
	ApplicationName  string
	ApplicationClass string
	rnk              int
}

func (this X11Window) buildWindowData() windowData {
	var wd = windowData{
		WindowId: uint32(this),
		Name:     this.GetName(),
		IconName: this.GetIconName(),
		State:    this.GetStates(),
	}
	wd.ApplicationName, wd.ApplicationClass = this.GetApplicationAndClass()
	return wd
}

func (this X11Window) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.buildWindowData())
}

func (this X11Window) DoDelete(w http.ResponseWriter, r *http.Request) {
	this.Close()
	respond.Accepted(w)
}

func (this X11Window) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if "" == action {
		this.RaiseAndFocus()
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

func (this X11Window) RaiseAndFocus() {
	proxy.Lock()
	defer proxy.Unlock()

	RaiseAndFocusWindow(proxy, uint32(this))
}

func (this X11Window) Close() {
	proxy.Lock()
	defer proxy.Unlock()

	CloseWindow(proxy, uint32(this))
}

func (this X11Window) GetName() string {
	proxy.Lock()
	defer proxy.Unlock()
	if name, err := GetName(proxy, uint32(this)); err != nil {
		return ""
	} else {
		return name
	}
}

func (this X11Window) GetApplicationAndClass() (string, string) {
	proxy.Lock()
	defer proxy.Unlock()
	return GetApplicationAndClass(proxy, uint32(this))
}

func (this X11Window) GetStates() WindowStateMask {
	proxy.Lock()
	defer proxy.Unlock()
	return GetStates(proxy, uint32(this))
}

func (this X11Window) GetIconName() string {
	proxy.Lock()
	defer proxy.Unlock()

	if name, err := GetIconName(proxy, uint32(this)); err != nil {
		return ""
	} else {
		return name
	}
}

func findNamedWindow(proxy Proxy, name string) (uint32, bool) {
	for _, wId := range GetStack(proxy) {
		if windowName, err := GetName(proxy, wId); err == nil && windowName == name {
			return wId, true
		}
	}
	return 0, false
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

var Windows = resource.MakeCollection[X11Window]()
var Monitors = resource.MakeCollection[*monitor.MonitorData]()
var proxy = MakeProxy()

func RaiseAndFocusNamedWindow(name string) bool {
	if w, ok := Windows.FindFirst(func(w X11Window) bool { return w.GetName() == name }); ok {
		w.RaiseAndFocus()
		return true
	} else {
		return false
	}
}

func Run() {

	var updateWindowList = func(p Proxy) {
		var wIds = GetStack(p)
		var xWins = make([]X11Window, len(wIds), len(wIds))
		for i := 0; i < len(wIds); i++ {
			xWins[i] = X11Window(wIds[i])
		}
		Windows.ReplaceWith(xWins)
	}

	var proxy = MakeProxy()
	SubscribeToEvents(proxy)
	updateWindowList(proxy)
	Monitors.ReplaceWith(GetMonitorDataList(proxy))

	for {
		event, _ := NextEvent(proxy)
		if event == DesktopStacking {
			updateWindowList(proxy)
		} else if event == DesktopGeometry {
			Monitors.ReplaceWith(GetMonitorDataList(proxy))
		}

	}
}

func getIconName(wId X11Window) string {
	proxy.Lock()
	defer proxy.Unlock()
	pixelArray, err := GetIcon(proxy, uint32(wId))
	if err != nil {
		log.Warn("Error converting x11 icon to pngs", err)
		return ""
	} else if name, err := icons.AddX11Icon(pixelArray); err != nil {
		log.Warn("Error adding icon:", err)
		return ""
	} else {
		return name
	}
}

func GetMonitors() []*monitor.MonitorData {
	return Monitors.GetAll()
}
