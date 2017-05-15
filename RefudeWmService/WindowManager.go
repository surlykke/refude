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
	"time"
)


var windows = make(map[xproto.Window]Window)
var display = Display{Screens: make([]Rect, 0)}
var iconHashes = make(map[uint64]bool)
var x  *xgbutil.XUtil

func getXConnection() (*xgbutil.XUtil, error) {
	var err error
	for i := 0; i < 5; i++ {
		if x, err := xgbutil.NewConn(); err == nil {
			return x, nil
		}
		time.Sleep(time.Second)
	}
	return nil, err
}

func getXgbConnection() (*xgb.Conn, error) {
	var err error
	for i := 0; i < 5; i++ {
		if conn, err := xgb.NewConn(); err == nil {
			return conn, nil
		}
		time.Sleep(time.Second)
	}
	return nil, err
}


func WmRun() {
	var err error
	if x, err = getXConnection(); err != nil {
		panic(err)
	}

	xwindow.New(x, x.RootWin()).Listen(xproto.EventMaskSubstructureNotify)
	updateWindows()

	conn, err := getXgbConnection()
	if err != nil {
		panic(err)
	}

	randr.Init(conn)
	buildDisplay(conn)

	for ;; {
		evt, err := x.Conn().WaitForEvent()
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
	tmp, err := ewmh.ClientListStackingGet(x)
	if err != nil {
		return
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
		}
	}

	for _,wId := range newWindowIds {
		if _, ok := windows[wId]; ok {
			windows[wId] = updateWindow(windows[wId])
		} else {
			windows[wId] = getWindow(xproto.Window(wId))
		}

		service.Map(fmt.Sprintf("/window/%d", wId), windows[wId])
	}

	mapWids(newWindowIds)
}

func getWindow(wId xproto.Window) Window {
	window := Window{}
	window.x = x
	window.Id = wId
	name, err := ewmh.WmNameGet(x, wId)
	if err != nil || len(name) == 0 {
		name,_ = icccm.WmNameGet(x, wId)
	}
	window.Name = name
	if rect, err := xwindow.New(x, wId).DecorGeometry(); err == nil {
		window.X = rect.X()
		window.Y = rect.Y()
		window.H = rect.Height()
		window.W = rect.Width()
	}

	if states, err := ewmh.WmStateGet(x, wId); err == nil {
		window.States = states
	}

	if iconArr, err := xprop.PropValNums(xprop.GetProperty(x, wId, "_NET_WM_ICON")); err == nil {
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

	window.Actions = make(map[string]Action)
	window.Actions["_default"] = Action{
		Name: window.Name,
		Comment: "Raise and focus",
		IconUrl: window.IconUrl,
		X: window.X,
		Y: window.Y,
		W: window.W,
		H: window.H,
	}

	return window
}

func updateWindow(window Window) Window {
	newWindow := Window{}
	newWindow.x = x
	newWindow.Id = window.Id
	name, err := ewmh.WmNameGet(x, newWindow.Id)
	if err != nil || len(name) == 0 {
		name,_ = icccm.WmNameGet(x, newWindow.Id)
	}
	newWindow.Name = name
	if rect, err := xwindow.New(x, newWindow.Id).DecorGeometry(); err == nil {
		newWindow.X = rect.X()
		newWindow.Y = rect.Y()
		newWindow.H = rect.Height()
		newWindow.W = rect.Width()
	}

	if states, err := ewmh.WmStateGet(x, newWindow.Id); err == nil {
		newWindow.States = states
	}

	newWindow.IconUrl = window.IconUrl

	newWindow.Actions = make(map[string]Action)
	newWindow.Actions["_default"] = Action{
		Name:    newWindow.Name,
		Comment: "Raise and focus",
		IconUrl: newWindow.IconUrl,
		X:       newWindow.X,
		Y:       newWindow.Y,
		W:       newWindow.W,
		H:       newWindow.H,
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

func mapWids(wIds []xproto.Window)  {
	res := make(common.StringList, len(wIds))
	for i,wId := range wIds {
		res[i] = fmt.Sprintf("window/%d", wId)
	}

	service.Map("/windows", res)
}

