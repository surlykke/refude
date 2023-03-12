package wayland

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/windows/monitor"
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
	Wid      uint64 `json:"id"`
	AppId    string `json:"app_id"`
	State    WindowStateMask
}

func (this *WaylandWindow) Actions() link.ActionList {
	return link.ActionList {
		{Name: "activate", Title: "Raise and focus", IconName: this.IconName},
		{Name: "close", Title: "Close", IconName: this.IconName},
	}
}

func (this *WaylandWindow) DoDelete(w http.ResponseWriter, r *http.Request) {
	close(this.Wid)
	respond.Accepted(w)
}

func (this *WaylandWindow) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if "" == action {
		activate(this.Wid)
		respond.Accepted(w)
	} else if "resize" == action {
		if width, err := strconv.ParseUint(requests.GetSingleQueryParameter(r, "width", ""), 10, 32); err != nil {
			respond.UnprocessableEntity(w, err)
		} else if height, err := strconv.ParseUint(requests.GetSingleQueryParameter(r, "height", ""), 10, 32); err != nil {
			respond.UnprocessableEntity(w, err)
		} else {
			setRectangle(this.Wid, 0, 0, uint32(width), uint32(height))
			respond.Accepted(w)
		}

	} else {
		respond.NotFound(w)
	}
}

type WaylandWindowManager struct {
	windows       *resource.Collection[*WaylandWindow]
	recentMap     map[uint64]uint32
	recentCount   uint32
	recentMapLock sync.Mutex
}

func MakeWaylandWindowManager() *WaylandWindowManager {
	var wwm = WaylandWindowManager{}
	wwm.windows = resource.MakeCollection[*WaylandWindow]()
	wwm.recentMap = make(map[uint64]uint32)
	return &wwm
}

var WM *WaylandWindowManager

func (this *WaylandWindowManager) GetPaths() []string {
	return this.windows.GetPaths()
}

func (this *WaylandWindowManager) handle_title(wId uint64, title string) {
	var ww = this.getCopy(wId)
	ww.Title = title
	this.windows.Put(&ww)
}

func (this *WaylandWindowManager) handle_app_id(wId uint64, app_id string) {
	var ww = this.getCopy(wId)
	ww.AppId = app_id
	this.windows.Put(&ww)
}

func (this *WaylandWindowManager) handle_output_enter(wId uint64, output uint64) {
}

func (this *WaylandWindowManager) handle_output_leave(wId uint64, output uint64) {
}

func (this *WaylandWindowManager) handle_state(wId uint64, state WindowStateMask) {
	var ww = this.getCopy(wId)
	ww.State = state
	this.windows.Put(&ww)
}

func (this *WaylandWindowManager) handle_done(wId uint64) {
}

func (this *WaylandWindowManager) handle_parent(wId uint64, parent uint64) {
}

func (this *WaylandWindowManager) handle_closed(wId uint64) {
	this.windows.Delete(strconv.FormatUint(wId, 10))
}

func (this *WaylandWindowManager) getCopy(wId uint64) WaylandWindow {
	var ww WaylandWindow
	if w, ok := this.windows.Get(strconv.FormatUint(wId, 10)); ok {
		ww = *w
	} else {
		ww.Wid = wId
	}
	return ww
}

func (this *WaylandWindowManager) RaiseAndFocusNamedWindow(name string) bool {
	if w, ok := this.windows.FindFirst(func(ww *WaylandWindow) bool { return ww.Title == name }); ok {
		activate(w.Wid)
		return true
	} else {
		return false
	}

}

func (this *WaylandWindowManager) GetMonitors() []*monitor.MonitorData {
	// TODO
	return []*monitor.MonitorData{}
}

func (this *WaylandWindowManager) Run() {
	setupAndRunAsWaylandClient()
}

func init() {
	if xdg.SessionType == "wayland" {
		WM = MakeWaylandWindowManager()
	}
}
