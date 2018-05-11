// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/icccm"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb"
	"time"
	"github.com/surlykke/RefudeServices/lib/icons"
	"github.com/surlykke/RefudeServices/lib/action"
	"github.com/surlykke/RefudeServices/lib/utils"
)


var windows = make(map[xproto.Window]*Window)
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

	service.Map("/display", &display, DisplayMediaType)
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

	for wId := range windows {
		if ! find(newWindowIds, wId) {
			delete(windows, wId)
			service.Unmap(fmt.Sprintf("/windows/%d", wId))
			service.Unmap(fmt.Sprintf("/actions/%d", wId))
		}
	}

	for i,wId := range newWindowIds {
		if _, ok := windows[wId]; ok {
			windows[wId] = updateWindow(windows[wId], i)
		} else {
			windows[wId] = getWindow(xproto.Window(wId), i)
		}
		var window = windows[wId]
		service.Map(fmt.Sprintf("/windows/%d", wId), window, WindowMediaType)
		if !  utils.Contains(window.States, "_NET_WM_STATE_ABOVE") { // TODO More that we won't offer as actions?
			var presentationHint string
			if utils.Among("_NET_WM_STATE_HIDDEN", windows[wId].States...) {
				presentationHint = "minimizedwindow"
			} else {
				presentationHint = "window"
			}
			var act= action.MakeAction(windows[wId].Name, "", windows[wId].IconName, presentationHint, MakeExecuter(wId))
			service.Map(fmt.Sprintf("/actions/%d", wId), act, action.ActionMediaType)
		}
	}
}

func getWindow(wId xproto.Window, stackingOrder int) *Window {
	window := Window{}
	window.x = x
	window.Id = wId
	name, err := ewmh.WmNameGet(x, wId)
	if err != nil || len(name) == 0 {
		name,_ = icccm.WmNameGet(x, wId)
	}
	window.Name = name
	window.RelevanceHint = -stackingOrder
	if rect, err := xwindow.New(x, wId).DecorGeometry(); err == nil {
		window.X = rect.X()
		window.Y = rect.Y()
		window.H = rect.Height()
		window.W = rect.Width()
	}

	if states, err := ewmh.WmStateGet(x, wId); err == nil {
		window.States = states
	} else {
		window.States = []string{}
	}

	if iconArr, err := xprop.PropValNums(xprop.GetProperty(x, wId, "_NET_WM_ICON")); err == nil {
		argbIcon := extractARGBIcon(iconArr)
		window.IconName = icons.SaveAsPngToSessionIconDir(argbIcon)
		fmt.Println("Setting iconname to: ", window.IconName)
	}

	window.Actions = make(map[string]Action)
	window.Actions["_default"] = Action{
		Name: window.Name,
		Comment: "Raise and focus",
	}

	return &window
}

func updateWindow(window *Window, stackOrder int) *Window {
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

	newWindow.IconName = window.IconName
	newWindow.IconUrl = window.IconUrl

	newWindow.Actions = make(map[string]Action)
	newWindow.Actions["_default"] = Action{
		Name:    newWindow.Name,
		Comment: "Raise and focus",
	}

	newWindow.RelevanceHint = -stackOrder
	return &newWindow
}

func find(windowIds []xproto.Window, windowId xproto.Window) bool {
	for _,wId := range windowIds {
		if wId == windowId {
			return true
		}
	}

	return false
}

func MakeExecuter(id xproto.Window) action.Executer {
	return func() {
		ewmh.ActiveWindowReq(x, xproto.Window(id))
	}
}


/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format (on 64bit system the 4 most
 * significant bytes are not used). After that it may repeat: again a width and height uint and then pixels and
 * so on...
 */
func extractARGBIcon(uints []uint) icons.Icon {
	res := make(icons.Icon, 0)
	for len(uints) >= 2 {
		width := int32(uints[0])
		height := int32(uints[1])
		fmt.Println("image dimensions: ", width, "x", height)

		uints = uints[2:]
		if len(uints) < int(width*height) {
			break
		}
		pixels := make([]byte, 4*width*height)
		for pos := int32(0); pos < width*height; pos++ {
			pixels[4*pos] = uint8((uints[pos] & 0xFF000000) >> 24)
			pixels[4*pos+1] = uint8((uints[pos] & 0xFF0000) >> 16)
			pixels[4*pos+2] = uint8((uints[pos] & 0xFF00) >> 8)
			pixels[4*pos+3] = uint8(uints[pos] & 0xFF)
		}
		res = append(res, icons.Img{Width: width, Height: height, Pixels: pixels})
		uints = uints[width*height:]
	}

	return res
}
