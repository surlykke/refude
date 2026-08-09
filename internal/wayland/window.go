// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package wayland

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/lib/resource"
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
		select {
		case upd := <-windowUpdates:
			var newWin = makeWindow(upd.wId)
			if oldWin, ok := resource.Get[*WaylandWindow](newWin.Path); ok {
				newWin.Base, newWin.AppId, newWin.State = oldWin.Base, oldWin.AppId, oldWin.State
			}

			if upd.title != "" {
				newWin.Title = upd.title
			}

			if upd.appId != "" {
				newWin.AppId = upd.appId
				if appTitle, appIconName, ok := applications.GetTitleAndIcon(upd.appId); ok {
					newWin.Summary = appTitle + " window"
					newWin.Icon = appIconName
				}
			}

			if upd.state > 0 {
				newWin.State = upd.state - 1
			}
			resource.Put(&newWin)
		case id := <-removals:
			resource.Remove(wId2Path(id))
		case _ = <-appEvents:
			for _, w := range resource.GetByPrefix[*WaylandWindow]("/window/") {
				var copy = *w
				if "" != copy.AppId {
					if appTitle, appIcon, ok := applications.GetTitleAndIcon(copy.AppId); ok {
						copy.Icon = appIcon
						copy.Summary = appTitle + " window"
						resource.Put(&copy)
					}
				}
			}
		}
	}
}

func watchAppCollections(sink chan struct{}) {
	var subscription = applications.AppEvents.Subscribe()
	for {
		subscription.Next()
		sink <- struct{}{}
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
	resource.Base
	Wid   uint64          `json:"-"`
	AppId string          `json:"app_id"`
	State WindowStateMask `json:"state"`
}


func makeWindow(wId uint64) WaylandWindow {
	var ww = WaylandWindow{
		Base: resource.MakeBase(wId2Path(wId), "", "", "", "Window"),
		Wid:  wId,
	}
	ww.AddAction("Focus", "", resource.OrgRefudeAction, "")
	ww.AddAction("close", "window-close", resource.OrgRefudeDelete, "")
	return ww
}

func (this *WaylandWindow) DoDelete() error {
	close(this.Wid)
	return nil
}

func (this *WaylandWindow) OmitFromSearch() bool {
	return strings.HasPrefix(this.Title, "Refude desktop") || ignoredWindows[this.AppId]
}

func (this *WaylandWindow) DoPost(action string) (bool, error) {
	if "" == action {
		activate(this.Wid)
		return true, nil
	} else {
		return false, nil
	}
}

func wId2Path(wId uint64) string {
	return fmt.Sprintf("/window/%d", wId)
}
