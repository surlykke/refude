package wayland

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

var windowUpdates = make(chan windowUpdate)
var removals = make(chan uint64)
var ignoredWindows map[string]bool

type windowUpdate struct {
	wId   uint64
	title string
	appId string
	state WindowStateMask
}

func Run(ignWin map[string]bool) {
	ignoredWindows = ignWin

	go setupAndRunAsWaylandClient()

	var appEvents = make(chan struct{})
	go watchAppCollections(appEvents)

	for {
		var publish = false
		select {
		case upd := <-windowUpdates:
			publish = (upd.title != "" && upd.title != "Refude Desktop") || upd.appId != ""
			//func MakeWindow(wId uint64, title, comment string, iconName icon.Name, appId string, state WindowStateMask) *WaylandWindow {

			var (
				title    string
				comment  string
				iconName icon.Name
				appId    string
				state    WindowStateMask
			)

			if w, ok := repo.Get[*WaylandWindow](path.Of("/window/", upd.wId)); ok {
				var self = w.Link()
				title, comment, iconName = self.Title, self.Comment, self.Icon
				appId, state = w.AppId, w.State
			}

			if upd.title != "" {
				title = upd.title
			}

			if upd.appId != "" {
				appId = upd.appId
				if appTitle, appIconName, ok := applications.GetTitleAndIcon(upd.appId); ok {
					comment = appTitle + " window"
					iconName = appIconName
				}
			}

			if upd.state > 0 {
				state = upd.state - 1
			}

			repo.Put(makeWindow(upd.wId, title, comment, iconName, appId, state))
		case id := <-removals:
			publish = true
			var path = path.Of("/window/", id)
			repo.Remove(path)
		case _ = <-appEvents:
			for _, w := range repo.GetList[*WaylandWindow]("/window/") {
				var self = w.Link()
				var title, comment, iconName = self.Title, self.Comment, self.Icon
				var appId, state = w.AppId, w.State

				if appTitle, appIconName, ok := applications.GetTitleAndIcon(appId); ok {
					comment = appTitle + " window"
					iconName = appIconName
					repo.Put(makeWindow(w.Wid, title, comment, iconName, appId, state))
				}
			}
		}
		if publish {
			watch.Publish("search", "")
		}
	}
}

func watchAppCollections(sink chan struct{}) {
	var subscription = applications.AppEvents.Subscribe()
	for {
		sink <- subscription.Next()
	}
}

type WindowStateMask uint8

const (
	MAXIMIZED = 1 << iota
	MINIMIZED
	ACTIVATED
	FULLSCREEN
)

func (wsm WindowStateMask) Is(other WindowStateMask) bool {
	return wsm&other == other
}

func (wsm WindowStateMask) toStringList() []string {
	var list = make([]string, 0, 4)
	if wsm&MAXIMIZED > 0 {
		list = append(list, "MAXIMIZED")
	}
	if wsm&MINIMIZED > 0 {
		list = append(list, "MINIMIZED")
	}
	if wsm&ACTIVATED > 0 {
		list = append(list, "ACTIVATED")
	}
	if wsm&FULLSCREEN > 0 {
		list = append(list, "FULLSCREEN")
	}
	return list
}

func (wsm WindowStateMask) String() string {
	return strings.Join(wsm.toStringList(), "|")

}

func (wsm WindowStateMask) MarshalJSON() ([]byte, error) {
	return json.Marshal(wsm.toStringList())
}

type WaylandWindow struct {
	resource.ResourceData
	Wid   uint64 `json:"-"`
	AppId string `json:"app_id"`
	State WindowStateMask
}

func makeWindow(wId uint64, title, comment string, iconName icon.Name, appId string, state WindowStateMask) *WaylandWindow {
	var ww = &WaylandWindow{
		ResourceData: *resource.MakeBase(path.Of("/window/", wId), title, comment, iconName, mediatype.Window),
		Wid:          wId,
		AppId:        appId,
		State:        state,
	}
	ww.AddAction("focus", title, "Focus window", iconName)
	//ww.AddAction("close", title, "Close window", "window-close")
	return ww
}

func (this *WaylandWindow) DoDelete(w http.ResponseWriter, r *http.Request) {
	close(this.Wid)
	respond.Accepted(w)
}

func (this *WaylandWindow) OmitFromSearch() bool {
	var self = this.Link()
	return strings.HasPrefix(self.Title, "Refude desktop") || ignoredWindows[this.AppId]
}

func (this *WaylandWindow) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if "focus" == action {
		activate(this.Wid)
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

var remembered atomic.Uint64

func RememberActive() {
	for _, w := range repo.GetList[*WaylandWindow]("/window/") {
		if w.State.Is(ACTIVATED) {
			remembered.Store(w.Wid)
			break
		}
	}
}

func ActivateRememberedActive() {
	if wId := remembered.Load(); wId > 0 {
		activate(wId)
	}
}
