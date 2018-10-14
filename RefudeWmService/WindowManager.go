// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"errors"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/surlykke/RefudeServices/lib"
	"log"
	"sort"
	"time"
)

const NET_WM_STATE_ABOVE = "_NET_WM_STATE_ABOVE"

type WindowWatcher struct {
	xutil         *xgbutil.XUtil
	xgbConn       *xgb.Conn
	defaultScreen *xproto.ScreenInfo

	windowList []*Window
}

func (watcher *WindowWatcher) ConnectToX11() {
	var err error
	if watcher.xutil, err = getXConnection(); err != nil {
		log.Fatal("No X connection", err)
	} else if watcher.xgbConn, err = getXgbConnection(); err != nil {
		log.Fatal("No xgb conn", err)
	} else if err := randr.Init(watcher.xgbConn); err != nil {
		panic(err)
	}
	watcher.defaultScreen = xproto.Setup(watcher.xgbConn).DefaultScreen(watcher.xgbConn)
}

func (w *WindowWatcher) Run() {
	w.ConnectToX11()

	var xEvents = make(chan xgb.Event)
	var ticks = make(chan struct{})

	go timer(ticks)
	go monitorEvents(w.xutil, xEvents)
	go w.watchDisplay()

	ticks <- struct{}{}
	var pending = true

	for {
		select {
		case <-ticks:
			w.GetAllWindowsAndActions()
			pending = false
		case e := <-xEvents:
			if ! pending {
				switch e.(type) {
				case
					xproto.CreateNotifyEvent,
					xproto.DestroyNotifyEvent,
					xproto.MapNotifyEvent,
					xproto.UnmapNotifyEvent,
					xproto.ConfigureNotifyEvent:
					ticks <- struct{}{}
					pending = true
				}
			}
		}
	}
}

func timer(ch chan struct{}) {
	for {
		<-ch
		time.Sleep(time.Millisecond * 100)
		ch <- struct{}{}
	}
}

func (w *WindowWatcher) watchDisplay() {

	var evtMask uint16 = randr.NotifyMaskScreenChange | randr.NotifyMaskCrtcChange | randr.NotifyMaskOutputChange | randr.NotifyMaskOutputProperty
	if err := randr.SelectInputChecked(w.xgbConn, w.defaultScreen.Root, evtMask).Check(); err != nil {
		panic(err)
	}

	for {
		resources, err := randr.GetScreenResources(w.xgbConn, w.defaultScreen.Root).Reply();
		if err != nil {
			panic(err)
		}
		var display Display
		display.Self = "/display"
		display.Mt = DisplayMediaType

		var rg = xwindow.RootGeometry(w.xutil)
		display.RootGeometry.X = rg.X()
		display.RootGeometry.Y = rg.Y()
		display.RootGeometry.W = uint(rg.Width())
		display.RootGeometry.H = uint(rg.Height())

		for _, crtc := range resources.Crtcs {
			if info, err := randr.GetCrtcInfo(w.xgbConn, crtc, 0).Reply(); err != nil {
				log.Fatal(err)
			} else if info.NumOutputs > 0 {
				var screen = Screen{X: int(info.X), Y: int(info.Y), W: uint(info.Width), H: uint(info.Height)}
				display.Screens = append(display.Screens, screen)
			}
		}

		sort.Sort(display.Screens)

		resourceCollection.Map(&display)

		if _, err := w.xgbConn.WaitForEvent(); err != nil {
			panic(err)
		}
	}
}

func monitorEvents(xutil *xgbutil.XUtil, sink chan xgb.Event) {
	xwindow.New(xutil, xutil.RootWin()).Listen(xproto.EventMaskSubstructureNotify)
	for {
		if evt, err := xutil.Conn().WaitForEvent(); err == nil {
			sink <- evt
		}
	}
}

/**
	Called, basically, whenever something happens. To paraphrase Ken Thomson: When you have no clue, use brute force.

 */
func (w *WindowWatcher) GetAllWindowsAndActions() {
	var newWindowList = make([]*Window, 0, 30)
	var resourcesToMap = make([]lib.Resource, 0, 60)
	if tmp, err := ewmh.ClientListStackingGet(w.xutil); err == nil {

		for _, wId := range tmp {
			if window, err := buildWindow(w.xutil, w.xgbConn, wId); err == nil {
				newWindowList = append(newWindowList, window)
				resourcesToMap = append(resourcesToMap, window)
				if (normal(window)) {
					var action = lib.MakeAction(
						lib.Standardizef("/window/%d/action", window.Id),
						window.Name,
						"Switch to this window",
						window.IconName,
						func() {
							ewmh.ActiveWindowReq(w.xutil, xproto.Window(window.Id))
						})
					lib.Relate(&window.AbstractResource, &action.AbstractResource)
					resourcesToMap = append(resourcesToMap, action)
				}
			}

		}

	}

	w.windowList = newWindowList
	resourceCollection.RemoveAndMap([]lib.StandardizedPath{"/window"}, resourcesToMap)
}

func buildWindow(xutil *xgbutil.XUtil, xgbConn *xgb.Conn, wId xproto.Window) (*Window, error) {
	if rect, err := xwindow.New(xutil, wId).DecorGeometry(); err != nil {
		return nil, err;
	} else if name, err := getWindowName(xutil, wId); err != nil {
		return nil, err;
	} else if states, err := ewmh.WmStateGet(xutil, wId); err != nil {
		return nil, err;
	} else {
		var window Window
		window.Id = wId
		window.Self = lib.Standardizef("/window/%d", wId)
		window.Mt = WindowMediaType
		window.Name = name
		window.Geometry.X = rect.X()
		window.Geometry.Y = rect.Y()
		window.Geometry.H = uint(rect.Height())
		window.Geometry.W = uint(rect.Width())
		window.States = states
		window.IconName = getIconName(xutil, wId)

		if tree, err := xproto.QueryTree(xgbConn, wId).Reply(); err == nil {
			window.parentId = tree.Parent
		}

		return &window, nil
	}
}

func getWindowName(xutil *xgbutil.XUtil, wId xproto.Window) (string, error) {
	if name, err := xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_VISIBLE_NAME")); err == nil {
		return name, nil
	} else if name, err = xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_NAME")); err == nil {
		return name, nil;
	} else if name, err = xprop.PropValStr(xprop.GetProperty(xutil, wId, "WM_NAME")); err == nil {
		return name, nil;
	} else {
		return "", errors.New("Neither '_NET_WM_TITLE_NAME', '_NET_WM_NAME' nor 'WM_NAME' set")
	}
}

func getIconName(xutil *xgbutil.XUtil, wId xproto.Window) string {
	/**
	  This doesn't work. openbox seems to put window titles in _NET_WM_VISIBLE_ICON_NAME and _NET_WM_ICON_NAME ??


	if name, err := xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_VISIBLE_ICON_NAME")); err == nil {
		return name
	} else if name, err := xprop.PropValStr(xprop.GetProperty(xutil, wId, "_NET_WM_ICON_NAME")); err == nil {
		return name
	} else
	*/

	if iconArr, err := xprop.PropValNums(xprop.GetProperty(xutil, wId, "_NET_WM_ICON")); err == nil {
		name := lib.SaveAsPngToSessionIconDir(extractARGBIcon(iconArr))
		return name
	} else {
		return ""
	}
}

func normal(window *Window) bool {
	return !lib.Contains(window.States, "_NET_WM_STATE_ABOVE")
}

/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format (on 64bit system the 4 most
 * significant bytes are not used). After that it may repeat: again a width and height uint and then pixels and
 * so on...
 */
func extractARGBIcon(uints []uint) lib.Icon {
	res := make(lib.Icon, 0)
	for len(uints) >= 2 {
		width := int32(uints[0])
		height := int32(uints[1])

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
		res = append(res, lib.Img{Width: width, Height: height, Pixels: pixels})
		uints = uints[width*height:]
	}

	return res
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
