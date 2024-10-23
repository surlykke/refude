package wayland

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/applications"
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
			var path = path.Of("/window/", upd.wId)
			var w WaylandWindow
			if tmp, ok := repo.Get[*WaylandWindow](path); ok {
				w = *tmp
			} else {
				w = *MakeWindow(upd.wId)
			}

			if upd.title != "" {
				w.Title = upd.title
			} else if upd.appId != "" {
				if app := applications.GetApp(upd.appId); app != nil {
					w.Comment, w.Icon = app.Title+" window", app.Icon
				}
				w.AppId = upd.appId
			} else if upd.state > 0 {
				w.State = upd.state - 1
			}

			repo.Put(&w)
		case id := <-removals:
			publish = true
			var path = path.Of("/window/", id)
			repo.Remove(path)
		case _ = <-appEvents:
			for _, w := range repo.GetList[*WaylandWindow]("/window/") {
				var copy = *w
				if app := applications.GetApp(copy.AppId); app != nil {
					copy.Comment, copy.Icon = app.Title+" window", app.Icon
				}
				repo.Put(&copy)
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

func MakeWindow(wId uint64) *WaylandWindow {
	var ww = &WaylandWindow{
		ResourceData: *resource.MakeBase(path.Of("/window/", wId), "", "", "", mediatype.Window),
		Wid:          wId,
	}
	ww.DefaultAction = "Focus window"
	ww.DeleteAction = "Close window"
	ww.DeleteIcon = "window-close"
	return ww
}

func (this *WaylandWindow) DoDelete(w http.ResponseWriter, r *http.Request) {
	close(this.Wid)
	respond.Accepted(w)
}

func (this *WaylandWindow) OmitFromSearch() bool {
	return strings.HasPrefix(this.Title, "Refude desktop") || ignoredWindows[this.AppId]
}

func (this *WaylandWindow) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "activate")
	if "activate" == action {
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
