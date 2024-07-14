package wayland

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var windowUpdates = make(chan windowUpdate)
var removals = make(chan uint64)

type windowUpdate struct {
	wId   uint64
	title string
	appId string
	state WindowStateMask
}

func Run() {

	go setupAndRunAsWaylandClient()

	var appIcons = make(map[string]string)
	var appIconChan = make(chan map[string]string)
	go watchAppCollections(appIconChan)

	for {
		select {
		case upd := <-windowUpdates:
			var path = fmt.Sprintf("/window/%d", upd.wId)
			var w WaylandWindow
			if tmp, ok := repo.Get[*WaylandWindow](path); ok {
				w = *tmp
			} else {
				w = *MakeWindow(upd.wId)
			}

			if upd.title != "" {
				w.Title = upd.title
			} else if upd.appId != "" {
				w.AppId = upd.appId
				w.Comment = upd.appId
				w.IconUrl = appIcons[w.AppId]
			} else if upd.state > 0 {
				w.State = upd.state - 1
			}

			repo.Put(&w)
		case id := <-removals:
			var path = fmt.Sprintf("/window/%d", id)
			repo.Remove(path)
		case appIcons = <-appIconChan:
			for _, w := range repo.GetList[*WaylandWindow]("/window/") {
				var copy = *w
				if iconUrl, ok := appIcons[w.AppId]; ok {
					copy.IconUrl = iconUrl
					repo.Put(&copy)
				}
			}
		}
	}
}

func watchAppCollections(sink chan map[string]string) {
	var subscription = applications.SubscribeToCollection()
	for {
		var collection = subscription.Next()
		var appIcons = make(map[string]string)
		for _, app := range collection.Apps {
			if app.IconUrl != "" {
				appIcons[app.DesktopId] = app.IconUrl
			}
		}
		sink <- appIcons
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

func (wsm WindowStateMask) MarshalJSON() ([]byte, error) {
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
	return json.Marshal(list)
}

type WaylandWindow struct {
	resource.ResourceData
	Wid   uint64 `json:"-"`
	AppId string `json:"app_id"`
	State WindowStateMask
}

func MakeWindow(wId uint64) *WaylandWindow {
	var ww = &WaylandWindow{
		ResourceData: *resource.MakeBase(fmt.Sprintf("/window/%d", wId), "", "", "", "window"),
		Wid:          wId,
	}
	ww.AddLink("", "Focus", "", relation.Action)
	ww.AddLink("", "Close", "", relation.Delete)
	return ww
}

func (this *WaylandWindow) DoDelete(w http.ResponseWriter, r *http.Request) {
	close(this.Wid)
	respond.Accepted(w)
}

func (this *WaylandWindow) OmitFromSearch() bool {
	return strings.HasPrefix(this.Title, "Refude Desktop")
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
