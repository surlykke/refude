package wayland

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
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
	Wid   uint64 `json:"id"`
	AppId string `json:"app_id"`
	State WindowStateMask
}

func MakeWindow(wId uint64) *WaylandWindow {
	return &WaylandWindow {
		BaseResource: resource.BaseResource {
			Id: strconv.FormatUint(wId, 10),
			Profile: "window",
		},
		Wid: wId,
	}
}

func (this *WaylandWindow) RelevantForSearch() bool {
	return this.Title != "Refude launcher" && this.Title != "Refude notifier"
}


func (this *WaylandWindow) Actions() link.ActionList {
	return link.ActionList{
		{Name: "activate", Title: "Raise and focus", IconName: this.IconName},
	}
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

var Windows = resource.MakeCollection[*WaylandWindow]("/window/")
var recentMap = make(map[uint64]uint32)
var recentCount uint32
var recentMapLock sync.Mutex

func getCopy(wId uint64) *WaylandWindow {
	var	copy WaylandWindow 
	var path = strconv.FormatUint(wId, 10)
	if w, ok := Windows.Get(path); ok {
		copy = *w	
	} else {
		copy = *MakeWindow(wId)
	}
	return &copy
}

func RaiseAndFocusNamedWindow(name string) bool {
	if w, ok := Windows.FindFirst(func(ww *WaylandWindow) bool { return ww.Title == name }); ok {
		activate(w.Wid)
		return true
	} else {
		return false
	}

}

func Run() {
	go watchApplications()
	setupAndRunAsWaylandClient()
}

func watchApplications() {
	var subscription = applications.Applications.Subscribe()
	for {
		if subscription.Next() == "/application/" {
			fmt.Println("updating icons...")
			for _, w := range Windows.GetAll() {
				var win = *w
				win.IconName = applications.GetIconName(win.AppId + ".desktop")
				Windows.Update(&win)
			}
		}
	}
}

func PurgeAndShow(applicationTitle string, focus bool) bool {
	if w := getAndPurge(applicationTitle); w == nil {
		return false
	} else {
		show(w.Wid)
		if focus {
			activate(w.Wid)
		}
		return true
	}
}


var rememberedActive uint64 = 0
var rememberedActiveLock sync.Mutex

func RememberActive() {
	if active, found := Windows.FindFirst(func(w *WaylandWindow) bool { return w.State.Is(ACTIVATED) }); found {
		rememberedActiveLock.Lock()
		rememberedActive = active.Wid
		rememberedActiveLock.Unlock()
	}
}

func ActivateRememberedActive() {
	rememberedActiveLock.Lock()
	var copy = rememberedActive
	rememberedActiveLock.Unlock()
	if (copy > 0) {
		activate(copy)
	}
}

func getAndPurge(applicationTitle string) *WaylandWindow {
	var result *WaylandWindow
	for _, w := range Windows.Find(func(w *WaylandWindow) bool { return w.Title == applicationTitle })  {
		if result == nil {
			result = w 
		} else {
			close(w.Wid)
		}
	}
	return result
}


