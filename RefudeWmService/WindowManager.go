/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"fmt"
	"hash/fnv"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/surlykke/RefudeServices/service"
	"net/http"
	"github.com/BurntSushi/xgb/randr"
	"github.com/surlykke/RefudeServices/common"
)


type WindowSet map[WId]*Window

func (wl WIdList) Data(r *http.Request) (int, string, []byte) {
	if r.Method == "GET" {
		paths := make([]string, len(wl), len(wl))
		for i,wId := range wl {
			paths[i] = fmt.Sprintf("window/%d", wId)
		}
		return common.GetJsonData(paths)
	} else {
		return http.StatusMethodNotAllowed, "", nil
	}
}


type WindowManager struct {
	wIds    WIdList
	windows WindowSet
	iconHashes map[uint64]bool
	x          *xgbutil.XUtil
}

func (wm *WindowManager) Run() {
	wm.wIds = make(WIdList, 0)
	service.Map("/windows", wm.wIds)
	wm.windows = make(WindowSet)
	wm.iconHashes = make(map[uint64]bool)
	var err error
	if wm.x, err = xgbutil.NewConn(); err != nil {
		panic(err)
	}
	xwindow.New(wm.x, wm.x.RootWin()).Listen(xproto.EventMaskSubstructureNotify)
	wm.updateWindows()

	for ;; {
		evt, err := wm.x.Conn().WaitForEvent()
		if err == nil {
			if scEvt, ok := evt.(randr.ScreenChangeNotifyEvent); ok {
				fmt.Println("Got screen change event: ", scEvt) // TODO update display resource
			} else {
				wm.updateWindows()
			}
		}
	}
}



func windowPath(wId WId) string {
	return fmt.Sprintf("/window/%d", wId)
}

func (wm *WindowManager) updateWindows() {
	var wIds WIdList
	var windows WindowSet = make(WindowSet)
	tmp, err := ewmh.ClientListStackingGet(wm.x)
	if err != nil {
		panic(err)
	}

	// Reverse order, so highest stacked comes first
	wIds = make(WIdList, len(tmp))
	for i,wId := range tmp {
		wIds[len(tmp) - 1 - i] = WId(wId)
	}

	for _,wId := range wIds {
		if window, err := wm.getWindow(xproto.Window(wId)); err == nil {
			windows[wId] = &window
		}
	}

	for wId,_ := range wm.windows {
		if _,ok := windows[wId]; !ok {
			service.Unmap(windowPath(wId))
		}
	}

	for wId, window := range windows {
		if oldWindow,ok := wm.windows[wId]; !ok {
			service.Map(windowPath(wId), window)
		} else {
			if ! oldWindow.Equal(window) {
				service.Remap(windowPath(wId), window)
			}
		}
	}

	if !(wIds.Equal(wm.wIds)) {
		service.Remap("/windows", wIds)
		wm.wIds = wIds
	}

	wm.windows = windows
}



func (wm *WindowManager) getWindow(wId xproto.Window) (Window, error) {
		window := Window{}
		window.x = wm.x
		window.Id = WId(wId)
		name, err := ewmh.WmNameGet(wm.x, wId)
		if err != nil || len(name) == 0 {
			name,_ = icccm.WmNameGet(wm.x, wId)
		}
		window.Name = name
		if rect, err := xwindow.New(wm.x, wId).DecorGeometry(); err == nil {
			window.X = rect.X()
			window.Y = rect.Y()
			window.H = rect.Height()
			window.W = rect.Width()
		}

		if states, err := ewmh.WmStateGet(wm.x, wId); err == nil {
			window.States = states
		}

		if iconArr, err := xprop.PropValNums(xprop.GetProperty(wm.x, wId, "_NET_WM_ICON")); err == nil {
			hash := fnv.New64a()
			for _,val := range iconArr {
				hash.Write([]byte{byte((val & 0xFF000000) >> 24), byte((val & 0xFF0000) >> 16), byte((val & 0xFF00) >> 8), byte(val & 0xFF)})
			}

			if !wm.iconHashes[hash.Sum64()] {
				if icon, err := MakeIcon(hash.Sum64(), iconArr); err == nil {
					iconUrl := fmt.Sprintf("/icon/%d", icon.hash)
					wm.iconHashes[icon.hash] = true
					service.Map(iconUrl, icon)
					window.IconUrl = ".." + iconUrl
				}
			}
		}

		window.Actions = make(map[string]Action)
		window.Actions["_default"] = Action{
			Name: window.Name,
			Comment: "Raise and focus",
			IconUrl: window.IconUrl,
		}

		return window, nil
}


