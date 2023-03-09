// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package x11

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/windows/monitor"
)

type X11Window uint32

func (this X11Window) Id() uint32 {
	return uint32(this)
}

func (this X11Window) Presentation() (title string, comment string, icon link.Href, profile string) {
	var name = WM.getName(this)
	var iconName = WM.getIconName(this)
	return name, "", link.IconUrl(iconName), "window"
}

func (this X11Window) Links(self, searchTerm string) link.List {
	var name, iconName string
	var err error
	name = WM.getName(this)
	iconName = WM.getIconName(this)

	if err == nil {
		var links = make(link.List, 0, 10)
		if rnk := searchutils.Match(searchTerm, name); rnk > -1 {
			links = append(links, link.Make(self, "Raise and focus", iconName, relation.DefaultAction))
			links = append(links, link.Make(self, "Close", iconName, relation.Delete))
		}
		return links
	}

	return link.List{}
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
		Name:     WM.getName(this),
		IconName: WM.getIconName(this),
		State:    WM.getStates(this),
	}
	wd.ApplicationName, wd.ApplicationClass = WM.getApplicationAndClass(this)
	return wd
}

func (this X11Window) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.buildWindowData())
}

func (this X11Window) DoDelete(w http.ResponseWriter, r *http.Request) {
	WM.CloseWindow(this)
	respond.Accepted(w)
}

func (this X11Window) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if "" == action {
		WM.RaiseAndFocusWindow(this)
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

type WindowGroup struct {
	Name    string
	Windows []uint32
}

func (this *WindowGroup) Id() string {
	return this.Name
}

func (this *WindowGroup) Presentation() (title string, comment string, icon link.Href, profile string) {
	var name = this.Name
	var iconName = ""
	if len(this.Windows) > 0 {
		iconName = WM.getIconName(X11Window(this.Windows[0]))
	}

	return name, "", link.IconUrl(iconName), "window"
}

func (this *WindowGroup) Links(self, searchTerm string) link.List {
	var links = make(link.List, 0, 20)
	for _, wId := range this.Windows {
		links = append(links, resource.LinkTo[uint32](X11Window(wId), "/window/", 0))
	}

	return links
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
		log.Warn("Error converting x11 icon to pngs", err)
		return "", err
	} else {
		return icons.AddX11Icon(pixelArray)
	}
}

type X11WindowManager struct {
	windows       *resource.Collection[uint32, X11Window]
	proxy         Proxy
	recentMap     map[uint32]uint32
	recentCount   uint32
	recentMapLock sync.Mutex
}

func makeX11WindowManager() *X11WindowManager {
	return &X11WindowManager{
		windows:   resource.MakePublishingCollection[uint32, X11Window]("/window/", "/search"),
		proxy:     MakeProxy(),
		recentMap: make(map[uint32]uint32),
	}
}

func (this *X11WindowManager) Lock() {
	this.proxy.Lock()
}

func (this *X11WindowManager) Unlock() {
	this.proxy.Unlock()
}

func (this *X11WindowManager) Search(sink chan link.Link, term string) {
	this.windows.Search(sink, func(xWin X11Window) int {
		var name = this.getName(xWin)
		var state = this.getStates(xWin)
		var appName, _ = this.getApplicationAndClass(xWin)
		if appName != "localhost__refude_html_launcher" &&
			appName != "localhost__refude_html_notifier" &&
			state&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) == 0 {
			if rnk := searchutils.Match(term, name); rnk > -1 {
				this.recentMapLock.Lock()
				defer this.recentMapLock.Unlock()
				var recentNess = this.recentCount - this.recentMap[uint32(xWin)]
				return int(recentNess) + rnk
			}
		}
		return -1
	})
}

func (this *X11WindowManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/window/group/") && len(r.URL.Path) > 14 {
		var groupName = r.URL.Path[14:]
		if group, found := this.buildGroup(groupName); !found {
			respond.NotFound(w)
		} else {
			respond.AsJson(w, resource.MakeWrapper[string](r.URL.Path, &group, ""))
		}
	} else if r.URL.Path == "/window/screenlayout" {
		if r.Method == "GET" {
			respond.AsJson(w, this.GetScreenLayoutFingerprint())
		} else {
			respond.NotAllowed(w)
		}
	} else if r.URL.Path == "/window/screen/" {
		if r.Method == "GET" {
			var wrappers []resource.Wrapper = make([]resource.Wrapper, 0, 4)
			for _, m := range this.GetMonitors() {
				wrappers = append(wrappers, monitor.MakeMonitorWrapper(m))
			}
			respond.AsJson(w, wrappers) 
		} else {
			respond.NotAllowed(w)
		}
	} else if strings.HasPrefix(r.URL.Path, "/window/screen/") {
		if m := this.GetMonitor(r.URL.Path[15:]); m != nil {
			respond.AsJson(w, monitor.MakeMonitorWrapper(m))
		} else {
			respond.NotFound(w)
		}
	} else {
		this.windows.ServeHTTP(w, r)
	}
}

func (this *X11WindowManager) GetPaths() []string {
	return this.windows.GetPaths()
}

func (this *X11WindowManager) RaiseAndFocusWindow(win X11Window) {
	this.Lock()
	defer this.Unlock()

	RaiseAndFocusWindow(this.proxy, uint32(win))
}

func (this *X11WindowManager) CloseWindow(win X11Window) {
	this.Lock()
	defer this.Unlock()

	CloseWindow(this.proxy, uint32(win))
}

func (this *X11WindowManager) RaiseAndFocusNamedWindow(name string) bool {
	this.Lock()
	defer this.Unlock()

	if wId, found := findNamedWindow(this.proxy, name); found {
		RaiseAndFocusWindow(this.proxy, wId)
		return true
	} else {
		return false
	}
}

func (this *X11WindowManager) Run() {

	var updateWindowList = func(p Proxy) {
		var wIds = GetStack(p)
		var xWins = make([]X11Window, len(wIds), len(wIds))
		for i := 0; i < len(wIds); i++ {
			xWins[i] = X11Window(wIds[i])
		}
		this.windows.ReplaceWith(xWins)
	}

	var proxy = MakeProxy()
	SubscribeToEvents(proxy)
	updateWindowList(proxy)
	for {
		event, _ := NextEvent(proxy)
		if event == DesktopStacking {
			updateWindowList(proxy)
		} else if event == ActiveWindow {
			if activeWindow, err := GetActiveWindow(proxy); err == nil {
				//this.recentMapLock.Lock()
				this.recentMap[activeWindow] = this.recentCount
				this.recentCount += 1
				//this.recentMapLock.Unlock()
			}
		}
	}
}

func (this *X11WindowManager) getName(wId X11Window) string {
	this.Lock()
	defer this.Unlock()
	if name, err := GetName(this.proxy, uint32(wId)); err != nil {
		return ""
	} else {
		return name
	}
}

func (this *X11WindowManager) getIconName(wId X11Window) string {
	this.Lock()
	defer this.Unlock()
	pixelArray, err := GetIcon(this.proxy, uint32(wId))
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

func (this *X11WindowManager) getStates(wId X11Window) WindowStateMask {
	this.Lock()
	defer this.Unlock()
	return GetStates(this.proxy, uint32(wId))
}

func (this *X11WindowManager) GetApplicationAndClass(wId X11Window) (string, string) {
	this.Lock()
	defer this.Unlock()
	return GetApplicationAndClass(this.proxy, uint32(wId))
}

func (this *X11WindowManager) GetScreenLayoutFingerprint() string {
	this.Lock()
	defer this.Unlock()
	var monitors = GetMonitorDataList(this.proxy)

	sort.Slice(monitors, func(i, j int) bool { return monitors[i].X < monitors[j].X })
	var fp = sha1.New()
	for _, m := range monitors {
		fp.Write([]byte(fmt.Sprintf(":%s:%d:%d:%d:%d", m.Name, m.X, m.Y, m.W, m.H)))
	}
	return hex.EncodeToString(fp.Sum(nil))
}

func (this *X11WindowManager) GetMonitors() []*monitor.MonitorData {
	this.Lock()
	defer this.Unlock()
	return GetMonitorDataList(this.proxy)
}

func (this *X11WindowManager) GetMonitor(name string) *monitor.MonitorData {
	for _, m := range GetMonitorDataList(this.proxy) {
		if name == m.Name {
			return m 
		}
	}
	return nil
}



func (this *X11WindowManager) getApplicationAndClass(wId X11Window) (string, string) {
	this.Lock()
	defer this.Unlock()
	return GetApplicationAndClass(this.proxy, uint32(wId))
}

func (this *X11WindowManager) buildGroup(name string) (WindowGroup, bool) {
	var windowList = make([]uint32, 0, 20)
	for _, wId := range this.windows.GetAll() {
		if applicationName, _ := this.getApplicationAndClass(wId); applicationName == name {
			windowList = append(windowList, uint32(wId))
		}
	}
	if len(windowList) == 0 {
		return WindowGroup{}, false
	} else {
		return WindowGroup{name, windowList}, true
	}
}


var WM *X11WindowManager

func init() {
	if "x11" == xdg.SessionType {
		WM = makeX11WindowManager()
	}
}
