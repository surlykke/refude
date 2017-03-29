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
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb"
	"github.com/surlykke/RefudeServices/common"
)


var windows = make(map[xproto.Window]Window)
var display = Display{Screens: make([]Rect, 0)}
var iconHashes = make(map[uint64]bool)


func WmRun() {
	var err error
	if xUtil, err = xgbutil.NewConn(); err != nil {
		panic(err)
	}

	xwindow.New(xUtil, xUtil.RootWin()).Listen(xproto.EventMaskSubstructureNotify)
	updateWindows()
	conn, err := xgb.NewConn()
	if err != nil {
		panic(err)
	}

	randr.Init(conn)
	buildDisplay(conn)

	for ;; {
		evt, err := xUtil.Conn().WaitForEvent()
		if err == nil {
			if scEvt, ok := evt.(randr.ScreenChangeNotifyEvent); ok {
				fmt.Println("Got screen change event: ", scEvt) // TODO update display resource
			} else {
				updateWindows()
			}
		}
	}
}


func buildDisplay(conn *xgb.Conn) {

	defaultScreen := xproto.Setup(conn).DefaultScreen(conn)
	display.W = defaultScreen.WidthInPixels
	display.H = defaultScreen.HeightInPixels

	// TODO add screens

	service.Map("/display", &display)
}

func updateWindows() {
	tmp, err := ewmh.ClientListStackingGet(xUtil)
	if err != nil {
		panic(err)
	}

	newWindowIds := make([]xproto.Window, len(tmp))
	// Reverse order, so highest stacked comes first
	for i, wId := range tmp {
		newWindowIds[len(tmp) - 1 -i] = wId
	}

	for wId,_ := range windows {
		if ! find(newWindowIds, wId) {
			delete(windows, wId)
			service.Unmap(fmt.Sprintf("/window/%d", wId))
			service.Unmap(fmt.Sprintf("/action/%d", wId))
		}
	}

	var windowPaths = make(common.StringList, 0, len(newWindowIds))
	var actionPaths = make(common.StringList, 0, len(newWindowIds))
	for _,wId := range newWindowIds {
		if _, ok := windows[wId]; ok {
			windows[wId] = updateWindow(windows[wId])
		} else {
			windows[wId] = getWindow(xproto.Window(wId))
		}

		windowPaths = append(windowPaths, fmt.Sprintf("window/%d", wId))
		service.Map(fmt.Sprintf("/window/%d", wId), windows[wId])
		if !common.Find(windows[wId].States, "_NET_WM_STATE_ABOVE") {
			service.Map(fmt.Sprintf("/action/%d", wId), windows[wId].Actions["_default"])
			actionPaths = append(actionPaths, fmt.Sprintf("action/%d", wId))
		}
	}

	service.Map("/windows", windowPaths)
	service.Map("/actions", actionPaths)
}

func getWindow(wId xproto.Window) Window {
	window := Window{}
	window.Id = wId
	name, err := ewmh.WmNameGet(xUtil, wId)
	if err != nil || len(name) == 0 {
		name,_ = icccm.WmNameGet(xUtil, wId)
	}
	window.Name = name
	if rect, err := xwindow.New(xUtil, wId).DecorGeometry(); err == nil {
		window.X = rect.X()
		window.Y = rect.Y()
		window.H = rect.Height()
		window.W = rect.Width()
	}

	if states, err := ewmh.WmStateGet(xUtil, wId); err == nil {
		window.States = states
	}

	if iconArr, err := xprop.PropValNums(xprop.GetProperty(xUtil, wId, "_NET_WM_ICON")); err == nil {
		hash := fnv.New64a()
		for _,val := range iconArr {
			hash.Write([]byte{byte((val & 0xFF000000) >> 24), byte((val & 0xFF0000) >> 16), byte((val & 0xFF00) >> 8), byte(val & 0xFF)})
		}

		iconUrl := fmt.Sprintf("/icon/%d", hash.Sum64())

		if !iconHashes[hash.Sum64()] {
			if icon, err := MakeIcon(hash.Sum64(), iconArr); err == nil {
				iconHashes[icon.hash] = true
				service.Map(iconUrl, icon)
			}
		}

		window.IconUrl = ".." + iconUrl
	}

	window.Actions = make(map[string]*Action)
	window.Actions["_default"] = &Action{
		winId: window.Id,
		Name: window.Name,
		Comment: "Raise and focus",
		IconUrl: window.IconUrl,
		X: window.X,
		Y: window.Y,
		W: window.W,
		H: window.H,
		States: window.States,
	}

	return window
}

func updateWindow(window Window) Window {
	newWindow := Window{}
	newWindow.Id = window.Id
	name, err := ewmh.WmNameGet(xUtil, newWindow.Id)
	if err != nil || len(name) == 0 {
		name,_ = icccm.WmNameGet(xUtil, newWindow.Id)
	}
	newWindow.Name = name
	if rect, err := xwindow.New(xUtil, newWindow.Id).DecorGeometry(); err == nil {
		newWindow.X = rect.X()
		newWindow.Y = rect.Y()
		newWindow.H = rect.Height()
		newWindow.W = rect.Width()
	}

	if states, err := ewmh.WmStateGet(xUtil, newWindow.Id); err == nil {
		newWindow.States = states
	}

	newWindow.IconUrl = window.IconUrl

	newWindow.Actions = make(map[string]*Action)
	newWindow.Actions["_default"] = &Action{
		winId:   newWindow.Id,
		Name:    newWindow.Name,
		Comment: "Raise and focus",
		IconUrl: newWindow.IconUrl,
		X:       newWindow.X,
		Y:       newWindow.Y,
		W:       newWindow.W,
		H:       newWindow.H,
		States:  newWindow.States,
	}

	return newWindow
}

func find(windowIds []xproto.Window, windowId xproto.Window) bool {
	for _,wId := range windowIds {
		if wId == windowId {
			return true
		}
	}

	return false
}

