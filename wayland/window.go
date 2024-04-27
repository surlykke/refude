package wayland

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/respond"
)

// Get current rect
//swaymsg -t get_tree | jq '.nodes[1].nodes[].floating_nodes[] | select(.name="org.refude.panel") | (.rect)'

// focus
// swaymsg '[title=org.refude.panel] focus'

// Move to
// swaymsg '[title=org.refude.panel] move absolute position 1200 0'

//
// swaymsg "[title=org.refude.panel] resize set width 200" (or height)

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
	resource.BaseResource
	Wid   uint64 `json:"-"`
	AppId string `json:"app_id"`
	State WindowStateMask
}


func MakeWindow(wId uint64) *WaylandWindow {
	var ww = &WaylandWindow{
		BaseResource: *resource.MakeBase(fmt.Sprintf("/window/%d", wId), "", "", "", "window"),
		Wid: wId,
	}
	ww.AddLink("", "Focus", "", relation.Action) 
	ww.AddLink("", "Close", "", relation.Delete) 
	return ww
}

func (this *WaylandWindow) GetIconUrl() string {
	return applications.GetIconUrl(this.AppId + ".desktop")
}

func (this *WaylandWindow) RelevantForSearch(term string) bool {
	return !strings.HasPrefix(this.Title, "Refude Desktop")
}

func (this *WaylandWindow) DoDelete(w http.ResponseWriter, r *http.Request) {
	close(this.Wid)
	respond.Accepted(w)
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

func retriewIconUrlsFromApps() {
	var windows = resourcerepo.GetTypedByPrefix[*WaylandWindow]("/window/")
	for _, w := range  windows{
		var copy = *w 
		var iconUrl = applications.GetIconUrl(copy.AppId + ".desktop")
		copy.IconUrl = iconUrl
		resourcerepo.Update(&copy)
	}
}

var recentMap = make(map[uint64]uint32)
var recentCount uint32
var recentMapLock sync.Mutex

func getCopy(wId uint64) *WaylandWindow {
	var copy WaylandWindow
	var path = fmt.Sprintf("/window/%d", wId)
	if w, ok := resourcerepo.GetTyped[*WaylandWindow](path); ok {
		copy = *w
	} else {
		copy = *MakeWindow(wId)
	}
	return &copy
}

func Run() {
	applications.AddListener(retriewIconUrlsFromApps) // FIXME this is racy
	setupAndRunAsWaylandClient()
}

var rememberedActive uint64 = 0
var rememberedActiveLock sync.Mutex

func RememberActive() {
	if active := resourcerepo.FindTypedUnderPrefix[*WaylandWindow]("/window/", func(w *WaylandWindow) bool {return w.State.Is(ACTIVATED) }); len(active) > 0 {
		rememberedActiveLock.Lock()
		rememberedActive = active[0].Wid
		rememberedActiveLock.Unlock()
	}
}

func ActivateRememberedActive() {
	rememberedActiveLock.Lock()
	var copy = rememberedActive
	rememberedActiveLock.Unlock()
	if copy > 0 {
		activate(copy)
	}
}
