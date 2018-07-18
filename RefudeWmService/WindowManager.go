// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"time"
	"fmt"
	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/ewmh"
	"github.com/BurntSushi/xgbutil/xprop"
	"github.com/BurntSushi/xgbutil/xwindow"
	"github.com/surlykke/RefudeServices/lib/action"
	"github.com/surlykke/RefudeServices/lib/icons"
	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/resource"
	"strings"
	"strconv"
	"github.com/surlykke/RefudeServices/lib/service"
	"log"
	"github.com/BurntSushi/xgb/randr"
	"sync"
	"sort"
)

const NET_WM_STATE_ABOVE = "_NET_WM_STATE_ABOVE"

var xutil *xgbutil.XUtil
var xgbConn *xgb.Conn
var setup *xproto.SetupInfo
var defaultScreen *xproto.ScreenInfo
var display *Display
var screensLock sync.Mutex

type Collection struct{}

func init() {
	var err error
	if xutil, err = getXConnection(); err != nil {
		log.Fatal("No X connection", err)
	} else if xgbConn, err = getXgbConnection(); err != nil {
		log.Fatal("No xgb conn", err)
	} else if err := randr.Init(xgbConn); err != nil {
		panic(err)
	}
	setup = xproto.Setup(xgbConn)
	defaultScreen = setup.DefaultScreen(xgbConn)
}

func maintainDisplay() {

	var evtMask uint16 = randr.NotifyMaskScreenChange | randr.NotifyMaskCrtcChange | randr.NotifyMaskOutputChange | randr.NotifyMaskOutputProperty
	if err := randr.SelectInputChecked(xgbConn, defaultScreen.Root, evtMask).Check(); err != nil {
		panic(err)
	}

	for {
		resources, err := randr.GetScreenResources(xgbConn, defaultScreen.Root).Reply();
		if err != nil {
			panic(err)
		}
		var newDisplay Display
		newDisplay.Self = "/display"
		newDisplay.Mt = DisplayMediaType

		var rg = xwindow.RootGeometry(xutil)
		newDisplay.RootGeometry.X = rg.X()
		newDisplay.RootGeometry.Y = rg.Y()
		newDisplay.RootGeometry.W = uint(rg.Width())
		newDisplay.RootGeometry.H = uint(rg.Height())

		for _, crtc := range resources.Crtcs {
			if info, err := randr.GetCrtcInfo(xgbConn, crtc, 0).Reply(); err != nil {
				log.Fatal(err)
			} else if info.NumOutputs > 0 {
				var screen = Screen{X: int(info.X), Y: int(info.Y), W: uint(info.Width), H: uint(info.Height)}
				newDisplay.Screens = append(newDisplay.Screens, screen)
			}
		}

		sort.Sort(newDisplay.Screens)

		screensLock.Lock()
		display = &newDisplay
		screensLock.Unlock()

		if _, err := xgbConn.WaitForEvent(); err != nil {
			panic(err)
		}
	}

}

func getDisplay() *Display {
	screensLock.Lock();
	defer screensLock.Unlock()
	return display
}

func (c *Collection) GetResource(path service.StandardizedPath) *resource.JsonResource {
	var res *resource.JsonResource = nil
	if path == "/links" {
		res = resource.MakeJsonResource(c.GetLinks())
	} else if path == "/display" {
		if d := getDisplay(); d != nil {
			res = resource.MakeJsonResource(d)
		}
	} else if strings.HasPrefix(string(path), "/windows/") {
		if u, err := strconv.ParseUint(string(path[9:]), 10, 32); err == nil {
			if window, _, err := buildWindowAndAction(xproto.Window(u)); err == nil && window != nil {
				res = resource.MakeJsonResource(window)
			}
		}
	} else if strings.HasPrefix(string(path), "/actions/") {
		if u, err := strconv.ParseUint(string(path[9:]), 10, 32); err == nil {
			if _, action, err := buildWindowAndAction(xproto.Window(u)); err == nil && action != nil {
				res = resource.MakeJsonResource(action)
			}
		}
	}

	if res != nil {
		res.EnsureReady()
	}

	return res
}

func (c *Collection) GetAll() []*resource.JsonResource {
	var allResources = []*resource.JsonResource{}
	if d := getDisplay(); d != nil {
		allResources = append(allResources, resource.MakeJsonResource(d))
	}
	if tmp, err := ewmh.ClientListStackingGet(xutil); err == nil {
		for _, wId := range tmp {
			if window, action, err := buildWindowAndAction(wId); err == nil {
				if window != nil {
					allResources = append(allResources, resource.MakeJsonResource(window))
				}
				if action != nil {
					allResources = append(allResources, resource.MakeJsonResource(action))
				}
			}
		}
	}

	for _, jsonRes := range allResources {
		jsonRes.EnsureReady()
	}

	return allResources
}

func (c *Collection) GetLinks() service.Links {
	var links = make(service.Links)
	links[DisplayMediaType] = []service.StandardizedPath{"/display"} // Small race here
	if tmp, err := ewmh.ClientListStackingGet(xutil); err == nil && len(tmp) > 0 {
		for _, wId := range tmp {
			links[WindowMediaType] = append(links[WindowMediaType], service.StandardizedPath(fmt.Sprintf("/windows/%d", wId)));
			if normalById(wId) {
				links[action.ActionMediaType] = append(links[action.ActionMediaType], service.StandardizedPath(fmt.Sprintf("/actions/%d", wId)))
			}
		}
	}

	return links
}

func buildWindowAndAction(wId xproto.Window) (*Window, *action.Action, error) {
	if window, err := buildWindow(wId); err != nil {
		return nil, nil, err
	} else if normal(window) {
		return window, buildAction(window), nil
	} else {
		return window, nil, nil
	}

}

func buildWindow(wId xproto.Window) (*Window, error) {
	if rect, err := xwindow.New(xutil, wId).DecorGeometry(); err != nil {
		return nil, err;
	} else if name, err := ewmh.WmNameGet(xutil, wId); err != nil {
		return nil, err;
	} else if states, err := ewmh.WmStateGet(xutil, wId); err != nil {
		return nil, err;
	} else if iconArr, err := xprop.PropValNums(xprop.GetProperty(xutil, wId, "_NET_WM_ICON")); err != nil {
		return nil, err;
	} else {
		var window Window
		window.Id = wId
		window.Self = fmt.Sprintf("/windows/%d", wId)
		window.Mt = WindowMediaType
		window.Name = name
		window.Geometry.X = rect.X()
		window.Geometry.Y = rect.Y()
		window.Geometry.H = uint(rect.Height())
		window.Geometry.W = uint(rect.Width())
		window.States = states
		argbIcon := extractARGBIcon(iconArr)
		window.IconName = icons.SaveAsPngToSessionIconDir(argbIcon)
		return &window, nil
	}
}

func buildAction(w *Window) *action.Action {
	return action.MakeAction(fmt.Sprintf("/actions/%d", w.Id), w.Name, "Switch to this window", w.IconName, func() {
		ewmh.ActiveWindowReq(xutil, xproto.Window(w.Id))
	})
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

func normal(w *Window) bool {
	return !utils.Contains(w.States, "_NET_WM_STATE_ABOVE")
}

func normalById(wId xproto.Window) bool {
	if states, err := ewmh.WmStateGet(xutil, wId); err == nil {
		return !utils.Contains(states, NET_WM_STATE_ABOVE)
	} else {
		return false
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
