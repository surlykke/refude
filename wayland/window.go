package wayland

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
)

var appIdAppDataChan = applications.MakeAppdIdAppDataChan()

var windowRepo = repo.MakeRepoWithFilter[*WaylandWindow](filter)
var repoRequests = repo.MakeAndRegisterRequestChan()

var windowUpdates = make(chan windowUpdate)
var removals = make(chan uint64)
var otherCommands = make(chan uint8)

type windowUpdate struct {
	wId   uint64
	title string
	appId string
	state WindowStateMask
}

func Run() {
	go setupAndRunAsWaylandClient()

	var appIdAppDataMap map[string]applications.AppData
	var rememberedActive uint64 = 0

	for {
		select {
		case req := <-repoRequests:
			windowRepo.DoRequest(req)
		case upd := <-windowUpdates:
			var path = fmt.Sprintf("/window/%d", upd.wId)
			var w WaylandWindow
			if tmp, ok := windowRepo.Get(path); ok {
				w = *tmp
			} else {
				w = *MakeWindow(upd.wId)
			}

			if upd.title != "" {
				w.Title = upd.title
			} else if upd.appId != "" {
				w.AppId = upd.appId
				w.Comment = upd.appId
				if appData, ok := appIdAppDataMap[upd.appId+".desktop"]; ok {
					w.IconUrl = appData.IconUrl
				}
			} else if upd.state > 0 {
				w.State = upd.state - 1
			}

			windowRepo.Put(&w)
		case id := <-removals:
			var path = fmt.Sprintf("/window/%d", id)
			windowRepo.Remove(path)
		case i := <-otherCommands:
			switch i {
			case 0:
				rememberedActive = 0
				for _, w := range windowRepo.GetAll() {
					if w.State&ACTIVATED > 0 {
						rememberedActive = w.Wid
						break
					}
				}
			case 1:
				if rememberedActive != 0 {
					activate(rememberedActive)
				}
			}
		case appIdAppDataMap = <-appIdAppDataChan:
			for _, ww := range windowRepo.GetAll() {
				if ww.AppId != "" {
					if appData, ok := appIdAppDataMap[ww.AppId+".desktop"]; ok {
						ww.IconUrl = appData.IconUrl
					}
				}
			}
		}
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

func (this *WaylandWindow) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "activate")
	if "activate" == action {
		activate(this.Wid)
		respond.Accepted(w)
	} else {
		respond.NotFound(w)
	}
}

func RememberActive() {
	otherCommands <- 0
}

func ActivateRememberedActive() {
	otherCommands <- 1
}

func filter(term string, ww *WaylandWindow) bool {
	return !strings.HasPrefix(ww.Title, "Refude Desktop")
}
