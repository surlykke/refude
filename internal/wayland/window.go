// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package wayland

import (
	"encoding/json"
	"strings"
	"sync/atomic"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/watch"
)

var WindowMap = entity.MakeMap[uint64, *WaylandWindow]()

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
				iconName icon.Name
				appId    string
				state    WindowStateMask
			)

			if w, ok := WindowMap.Get(upd.wId); ok {
				title, iconName = w.Title, w.Icon
				appId, state = w.AppId, w.State
			}

			if upd.title != "" {
				title = upd.title
			}

			if upd.appId != "" {
				appId = upd.appId
				if _, appIconName, ok := applications.GetTitleAndIcon(upd.appId); ok {
					iconName = appIconName
				}
			}

			if upd.state > 0 {
				state = upd.state - 1
			}

			WindowMap.Put(upd.wId, makeWindow(upd.wId, title, iconName, appId, state))
		case id := <-removals:
			publish = true
			WindowMap.Remove(id)
		case _ = <-appEvents:
			for _, w := range WindowMap.GetAll() {
				var title, iconName = w.Title, w.Icon
				var appId, state = w.AppId, w.State

				if _, appIconName, ok := applications.GetTitleAndIcon(appId); ok {
					iconName = appIconName
					WindowMap.Put(w.Wid, makeWindow(w.Wid, title, iconName, appId, state))
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
	entity.Base
	Wid   uint64 `json:"-"`
	AppId string `json:"app_id"`
	State WindowStateMask
}

func makeWindow(wId uint64, title string, iconName icon.Name, appId string, state WindowStateMask) *WaylandWindow {
	var ww = &WaylandWindow{
		Base:  *entity.MakeBase(title, appId+" "+"window", iconName, entity.Window),
		Wid:   wId,
		AppId: appId,
		State: state,
	}
	ww.AddAction("", "Focus", "")
	//ww.AddAction("close", title, "Close window", "window-close")
	return ww
}

func (this *WaylandWindow) DoDelete() response.Response {
	close(this.Wid)
	return response.Accepted()
}

func (this *WaylandWindow) OmitFromSearch() bool {
	return strings.HasPrefix(this.Title, "Refude desktop") || ignoredWindows[this.AppId]
}

func (this *WaylandWindow) DoPost(action string) response.Response {
	if "" == action {
		activate(this.Wid)
		return response.Accepted()
	} else {
		return response.NotFound()
	}
}

var remembered atomic.Uint64

func RememberActive() {
	for _, w := range WindowMap.GetAll() {
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
