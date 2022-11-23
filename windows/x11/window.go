// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package x11

import (
	"encoding/json"
	"net/http"
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
		/*var panes = collectPanes()
		for _, pane := range panes {
			if pane.XWinId == this.Id() {
				var title = pane.CurrentCommand + ":" + pane.CurrentDirectory
				if rnk := searchutils.Match(searchTerm, title, "pane", "tmux"); rnk > -1 {
					links = append(links, link.MakeRanked("/tmux/"+pane.PaneId, title, "", "tmux", rnk))
				}
			}
		}*/
		return links
	}

	return link.List{}
}

func (this X11Window) MarshalJSON() ([]byte, error) {
	var jsonData struct {
		WindowId uint32
		Name     string
		IconName string `json:",omitempty"`
		State    WindowStateMask
	}

	jsonData.WindowId = uint32(this)
	jsonData.Name = WM.getName(this)
	jsonData.IconName = WM.getIconName(this)
	jsonData.State = WM.getStates(this)

	return json.Marshal(jsonData)
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



func (this *X11WindowManager) Search(term string) link.List {
	return this.windows.ExtractLinks(func(xWin X11Window) int {
		var name = this.getName(xWin)
		var state = this.getStates(xWin)
		if state&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) == 0 {
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
	this.windows.ServeHTTP(w, r)
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

func (this *X11WindowManager) ResizePanel(newWidth, newHeight uint32) bool {
	this.Lock()
	defer this.Unlock()

	if wId, found := findNamedWindow(this.proxy, "org.refude.panel"); found {
		Resize(this.proxy, wId, newWidth, newHeight)
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

var WM *X11WindowManager

func init() {
	if "x11" == xdg.SessionType {
		WM = makeX11WindowManager()
	}
}


