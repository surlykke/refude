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

var display = Display{Screens: make([]Rect, 0)}
var windows = make(map[xproto.Window]*Window)
var xutil *xgbutil.XUtil

func Run() {
	var err error
	if xutil, err = getXConnection(); err != nil {
		panic(err)
	}

	xwindow.New(xutil, xutil.RootWin()).Listen(xproto.EventMaskSubstructureNotify)

	if err != nil {
		return
	}

	getWindows()

	conn, err := getXgbConnection()
	if err != nil {
		panic(err)
	}

	randr.Init(conn)
	buildDisplay(conn)

	var serverEvents = make(chan xgb.Event)
	go watchServer(serverEvents)

	for {
		select {
		case evt := <-serverEvents:

			switch e := evt.(type) {
			case xproto.CreateNotifyEvent:
				fmt.Println(evt)
				getWindows()
			case xproto.ConfigureNotifyEvent:
				if w, ok := windows[e.Window]; ok {
					copy := *w
					copy.X = int(e.X)
					copy.Y = int(e.Y)
					copy.H = int(e.Height)
					copy.W = int(e.Width)
					mapWindow(&copy)
				}
			case xproto.DestroyNotifyEvent:
				fmt.Println(evt)
				delete(windows, e.Window)
				service.Unmap(fmt.Sprintf("/windows/%d", e.Window))
				service.Unmap(fmt.Sprintf("/actions/%d", e.Window))
			case xproto.MapNotifyEvent:
				if w, ok := windows[e.Window]; ok {
					copy := *w
					copy.States = utils.Filter(copy.States, func(state string) bool { return state != "_NET_WM_STATE_HIDDEN" })
					mapWindow(&copy)
				}
			case xproto.UnmapNotifyEvent:
				if w, ok := windows[e.Window]; ok {
					copy := *w
					if !utils.Contains(copy.States, "_NET_WM_STATE_HIDDEN") {
						copy.States = append(copy.States, "_NET_WM_STATE_HIDDEN")
						mapWindow(&copy)
					}
				}
			case randr.ScreenChangeNotifyEvent:
				fmt.Println("Got screen change event: ", e) // TODO update display resource
			default:
			}
		}
	}
}

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

func buildDisplay(conn *xgb.Conn) {

	defaultScreen := xproto.Setup(conn).DefaultScreen(conn)
	display.W = defaultScreen.WidthInPixels
	display.H = defaultScreen.HeightInPixels
	// TODO add screens

	service.Map("/display", &display, DisplayMediaType)
}


func watchServer(evts chan xgb.Event) {
	for ; ; {
		if evt, err := xutil.Conn().WaitForEvent(); err == nil {
			evts <- evt
		}
	}
}

func getWindows() {
	if tmp, err := ewmh.ClientListStackingGet(xutil); err == nil {
		for _, wId := range tmp {
			if _, ok := windows[wId]; !ok {
				if rect, err := xwindow.New(xutil, wId).DecorGeometry(); err == nil {
					w := getWindow(wId, rect.X(), rect.Y(), rect.Height(), rect.Width())
					mapWindow(w)
					windows[wId] = w
				}
			}
		}
	}
}

func mapWindow(w *Window) {
	windows[w.Id] = w
	service.Map(fmt.Sprintf("/windows/%d", w.Id), w, WindowMediaType)
	if normal(w) {
		var presentationHint string
		if utils.Contains(w.States, "_NET_WM_STATE_HIDDEN") {
			presentationHint = "minimizedwindow"
		} else {
			presentationHint = "window"
		}
		var act = action.MakeAction(w.Name, "", w.IconName, presentationHint, MakeExecuter(w.Id))
		service.Map(fmt.Sprintf("/actions/%d", w.Id), act, action.ActionMediaType)
	}
}

func normal(w *Window) bool {
	return ! utils.Contains(w.States, "_NET_WM_STATE_ABOVE")
}



func getWindow(id xproto.Window, x int, y int, h int, w int) *Window {
	window := Window{}
	window.Id = id

	name, err := ewmh.WmNameGet(xutil, window.Id)
	if err != nil || len(name) == 0 {
		name, _ = icccm.WmNameGet(xutil, window.Id)
	}
	window.Name = name

	window.X = x
	window.Y = y
	window.H = h
	window.W = w

	if states, err := ewmh.WmStateGet(xutil, window.Id); err == nil {
		window.States = states
	} else {
		window.States = []string{}
	}

	if iconArr, err := xprop.PropValNums(xprop.GetProperty(xutil, id, "_NET_WM_ICON")); err == nil {
		argbIcon := extractARGBIcon(iconArr)
		window.IconName = icons.SaveAsPngToSessionIconDir(argbIcon)
		fmt.Println("Setting iconname to: ", window.IconName)
	}

	return &window
}

func MakeExecuter(id xproto.Window) action.Executer {
	return func() {
		ewmh.ActiveWindowReq(xutil, xproto.Window(id))
	}
}

/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format (on 64bit system the 4 most
 * significant bytes are not used). After that it may repeat: again a width and height uint and then pixels and
 * so on...
 */
func
extractARGBIcon(uints []uint) icons.Icon {
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
