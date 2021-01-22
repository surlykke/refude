// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Bounds struct {
	X, Y int32
	W, H uint32
}

type Geometry uint32

func (g Geometry) MarshalJSON() ([]byte, error) {
	return respond.ToJson(GetBounds(dataConnection, uint32(g))), nil
}

type Window struct {
	respond.Links `json:"_links"`
	Id            uint32
	Parent        uint32
	Name          string
	IconName      string `json:",omitempty"`
	Geometry      Geometry
	State         WindowStateMask
}

// Caller ensures thread safety
func BuildWindow(c *Connection, wId uint32) *Window {
	var win = &Window{Id: wId}
	win.Parent, _ = GetParent(c, wId)
	win.Name, _ = GetName(c, wId)
	win.IconName, _ = GetIconName(c, wId)
	if win.Parent > 0 {
		win.Geometry = Geometry(win.Parent)
	} else {
		win.Geometry = Geometry(win.Id)
	}
	win.State, _ = GetState(c, wId)

	BuildLinks(win)
	return win
}

// Caller ensures thread safety
func BuildLinks(win *Window) {
	var monitorList = monitors.Load().([]*MonitorData)
	var href = fmt.Sprintf("/window/%d", win.Id)
	var actionPrefix = href + "?action="
	win.Links = make(respond.Links, 0, 5+len(monitorList))
	win.Links = win.Links.Add(href, win.Name, icons.IconUrl(win.IconName), respond.Self, "/profile/window", respond.Hints{"states": win.State})

	win.Links = win.Links.Add(href+"/screenshot", "Screenshot of "+win.Name, "", respond.Related, "/profile/window-screenshot", nil)
	//wd.Links = wd.Links.Add(actionPrefix+"raise", "Raise and focus", "", respond.Action, "", nil)

	if win.State&(HIDDEN|MAXIMIZED_HORZ|MAXIMIZED_VERT) != 0 {
		win.Links = win.Links.Add(actionPrefix+"restore", "Restore", "", respond.Action, "", nil)
	} else {
		win.Links = win.Links.Add(actionPrefix+"minimize", "Minimize", "", respond.Action, "", nil)
		win.Links = win.Links.Add(actionPrefix+"maximize", "Maximize", "", respond.Action, "", nil)
	}

	for _, m := range monitorList {
		win.Links = win.Links.Add(actionPrefix+"move&monitor="+m.Name, "Move to monitor "+m.Name, "", respond.Action, "", nil)
	}
}

func (win Window) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, win)
	} else if r.Method == "POST" {
		var action = requests.Action(r)
		switch action {
		case "":
			RaiseAndFocusWindow(dataConnection, win.Id)
			respond.Accepted(w)
		case "restore":
			if win.State&HIDDEN > 0 {
				RemoveStates(dataConnection, win.Id, HIDDEN)
			} else {
				RemoveStates(dataConnection, win.Id, MAXIMIZED_HORZ|MAXIMIZED_VERT)
			}
		case "maximize":
			fmt.Println("maximizing")
			AddStates(dataConnection, win.Id, MAXIMIZED_VERT|MAXIMIZED_HORZ)
			respond.Accepted(w)
		case "minimize":
			AddStates(dataConnection, win.Id, HIDDEN)
			respond.Accepted(w)
		case "move":
			monitorName := requests.GetSingleQueryParameter(r, "monitor", "")
			for _, m := range monitors.Load().([]*MonitorData) {
				if m.Name == monitorName {
					var maximized = win.State & (MAXIMIZED_HORZ | MAXIMIZED_VERT)
					RemoveStates(dataConnection, win.Id, maximized)
					var bounds = GetBounds(dataConnection, win.Id)
					var marginW, marginH = m.W / 10, m.H / 10
					var newX, newY = m.X + int32(marginW), m.Y + int32(marginH)
					var newW, newH = bounds.W, bounds.H
					if newW == 0 || newW > m.W {
						newW = m.W - 2*marginW
					}
					if newH == 0 || newH > m.H {
						newH = m.H - 2*marginH
					}
					SetBounds(dataConnection, win.Id, newX, newY, newW, newH)
					AddStates(dataConnection, win.Id, maximized)
					RaiseAndFocusWindow(dataConnection, win.Id)
					respond.Accepted(w)
					return
				}
			}
			respond.UnprocessableEntity(w, fmt.Errorf("No such monitor '%s'", monitorName))
		default:
			respond.UnprocessableEntity(w, fmt.Errorf("Unknown action %s", action))
		}
	} else if r.Method == "DELETE" {
		dataConnection.CloseWindow(win.Id)
		respond.Accepted(w)
	} else {
		respond.NotAllowed(w)
	}
}

type ScreenShot uint32

func (ss ScreenShot) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var downscaleS = requests.GetSingleQueryParameter(r, "downscale", "1")
		if len(downscaleS) != 1 || downscaleS[0] < '1' || downscaleS[0] > '5' {
			respond.UnprocessableEntity(w, fmt.Errorf("downscale should be a number between 1 and 5 (inclusive)"))

		} else {
			var downscale = downscaleS[0] - '0'
			if bytes, err := getScreenshot(uint32(ss), downscale); err == nil {
				w.Header().Set("Content-Type", "image/png")
				w.Write(bytes)
			} else {
				respond.ServerError(w, err)
			}
		}
	} else {
		respond.NotAllowed(w)
	}
}

func getScreenshot(wId uint32, downscale byte) ([]byte, error) {
	return GetScreenshotAsPng(dataConnection, wId, downscale)
}

var dataConnection = MakeDisplay()

// Pulling icons from X11 (as GetIconName below does) is somewhat costly. For example 'Visual Studio Code' has a
// window icon of size 1024x1024, so it contains ~ 4 Mb. Hence the caching.
// TODO: Update cache on icon change event (?). Purge cache on window close
var iconNameCache = make(map[uint32]string)
var iconNameCacheLock sync.Mutex

func getIconNameFromCache(wId uint32) (string, bool) {
	iconNameCacheLock.Lock()
	defer iconNameCacheLock.Unlock()
	name, ok := iconNameCache[wId]
	return name, ok
}

func setIconNameInCache(wId uint32, name string) {
	iconNameCacheLock.Lock()
	defer iconNameCacheLock.Unlock()
	iconNameCache[wId] = name
}

func GetIconName(c *Connection, wId uint32) (string, error) {
	if name, ok := getIconNameFromCache(wId); ok {
		return name, nil
	} else {
		pixelArray, err := GetIcon(c, wId)
		if err != nil {
			return "", err
		}
		/*
		 * Icons retrieved from the X-server (EWMH) will come as arrays of uint32. There will be first two ints giving
		 * width and height, then width*height uints each holding a pixel in ARGB format.
		 * After that it may repeat: again a width and height uint and then pixels and
		 * so on...
		 */
		var images = []image.ARGBImage{}
		for len(pixelArray) >= 2 {
			width := pixelArray[0]
			height := pixelArray[1]

			pixelArray = pixelArray[2:]
			if len(pixelArray) < int(width*height) {
				break
			}
			pixels := make([]byte, 4*width*height)
			for pos := uint32(0); pos < width*height; pos++ {
				pixels[4*pos] = uint8((pixelArray[pos] & 0xFF000000) >> 24)
				pixels[4*pos+1] = uint8((pixelArray[pos] & 0xFF0000) >> 16)
				pixels[4*pos+2] = uint8((pixelArray[pos] & 0xFF00) >> 8)
				pixels[4*pos+3] = uint8(pixelArray[pos] & 0xFF)
			}
			images = append(images, image.ARGBImage{Width: width, Height: height, Pixels: pixels})
			pixelArray = pixelArray[width*height:]
		}

		var icon = image.ARGBIcon{Images: images}
		var iconName = icons.AddARGBIcon(icon)
		setIconNameInCache(wId, iconName)
		return iconName, nil
	}
}

var noBounds = &Bounds{0, 0, 0, 0}

func GetBounds(c *Connection, wId uint32) *Bounds {
	// TODO Perhaps some caching
	if x, y, w, h, err := dataConnection.GetGeometry(wId); err != nil {
		return noBounds
	} else {
		return &Bounds{x, y, w, h}
	}
}
