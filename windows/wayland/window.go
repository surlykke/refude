package wayland

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os/exec"
	"strconv"
	"sync"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
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
	Wid      uint64 `json:"id"`
	Title    string `json:"title"`
	AppId    string `json:"app_id"`
	State    WindowStateMask
	IconName string `json:",omitempty"`
}

func (this *WaylandWindow) Id() uint64 {
	return this.Wid
}

func (this *WaylandWindow) Presentation() (string, string, link.Href, string) {
	var iconUrl = link.IconUrl(applications.GetIconName(this.AppId))
	return this.Title, "", iconUrl, "window"
}

func (this *WaylandWindow) Links(self, term string) link.List {
	var links = make(link.List, 0, 10)
	if rnk := searchutils.Match(term, this.Title); rnk > -1 {
		links = append(links, link.Make(self, "Raise and focus", this.IconName, relation.DefaultAction))
		links = append(links, link.Make(self, "Close", this.IconName, relation.Delete))
	}

	return links
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



type WaylandWindowManager struct{
	windows *resource.Collection[uint64, *WaylandWindow]
	recentMap     map[uint64]uint32
	recentCount   uint32
	recentMapLock sync.Mutex
}

func MakeWaylandWindowManager() *WaylandWindowManager {
	var wwm = WaylandWindowManager{}
	wwm.windows = resource.MakeCollection[uint64, *WaylandWindow]("/window/")
	wwm.recentMap = make(map[uint64]uint32)
	return &wwm
}


var WM *WaylandWindowManager

func (this *WaylandWindowManager) Search(term string) link.List {
	return this.windows.ExtractLinks(func(wWin *WaylandWindow) int {
		if wWin.Title != "org.refude.panel"  {
			if rnk := searchutils.Match(term, wWin.Title); rnk > -1 {
				this.recentMapLock.Lock()
				defer this.recentMapLock.Unlock()
				var recentNess = this.recentCount - this.recentMap[wWin.Wid]
				return int(recentNess) + rnk
			}
		}
		return -1
	})
}

func (this *WaylandWindowManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	this.windows.ServeHTTP(w, r)
}

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
	this.windows.Delete(wId)	
}

func (this *WaylandWindowManager) getCopy(wId uint64) WaylandWindow {
	var ww WaylandWindow
	if w := this.windows.Get(wId); w != nil {
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

func (this *WaylandWindowManager) ResizePanel(newWidth, newHeight uint32) bool{
	// Can't figure out how to use foreign protocol set_rectangle
	var cmd = fmt.Sprintf("[title=org.refude.panel] resize set width %d;", newWidth) +
	          fmt.Sprintf("[title=org.refude.panel] resize set height %d;", newHeight) +
	          "[title=org.refude.panel] move absolute position 0 0"
	if err := exec.Command("swaymsg", cmd).Run(); err != nil {
		log.Warn(err)
		return false
	} else {
		return true
	}
}

func (this *WaylandWindowManager) Run() {
	setupAndRunAsWaylandClient()
}

func init() {
	if xdg.SessionType == "wayland" {
		WM = MakeWaylandWindowManager()
	}
}
