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

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/slice"

	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib/image"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
)

type WindowData struct {
	respond.Links `json:"_links"`
	Id            uint32
	Parent        uint32
	X, Y          int32
	W, H          uint32
	Name          string
	IconName      string `json:",omitempty"`
	States        []string
}

type Window uint32

func (win Window) ToData() *WindowData {
	var id = uint32(win)
	var wd = WindowData{Id: id}
	dataMutex.Lock()
	if parent, err := dataConnection.GetParent(id); err == nil {
		if parent != 0 {
			wd.X, wd.Y, wd.W, wd.H, err = dataConnection.GetGeometry(parent)
		} else {
			wd.X, wd.Y, wd.W, wd.H, err = dataConnection.GetGeometry(id)
		}
	}
	if name, err := dataConnection.GetName(id); err == nil {
		wd.Name = name
	}
	if iconName, err := GetIconName(id); err == nil {
		wd.IconName = iconName
	}
	if states, err := dataConnection.GetState(id); err == nil {
		wd.States = states
	}
	dataMutex.Unlock()

	var monitorList = monitors.Load().([]*Monitor)
	var href = fmt.Sprintf("/window/%d", id)
	var actionPrefix = href + "?action="
	var state map[string]string
	if slice.Contains(wd.States, "_NET_WM_STATE_HIDDEN") {
		state = map[string]string{"state": "minimized"}
	}
	wd.Links = make(respond.Links, 0, 5+len(monitorList))
	wd.Links = wd.Links.Add(href, wd.Name, icons.IconUrl(wd.IconName), respond.Self, "/profile/window", state)

	wd.Links = wd.Links.Add(href+"/screenshot", "Screenshot of "+wd.Name, "", respond.Related, "/profile/window-screenshot", nil)
	//wd.Links = wd.Links.Add(actionPrefix+"raise", "Raise and focus", "", respond.Action, "", nil)

	if slice.Contains(wd.States, NET_WM_STATE_HIDDEN) || (slice.Contains(wd.States, NET_WM_STATE_MAXIMIZED_HORZ) && slice.Contains(wd.States, NET_WM_STATE_MAXIMIZED_VERT)) {
		wd.Links = wd.Links.Add(actionPrefix+"restore", "Restore", "", respond.Action, "", nil)
	} else {
		wd.Links = wd.Links.Add(actionPrefix+"minimize", "Minimize", "", respond.Action, "", nil)
		wd.Links = wd.Links.Add(actionPrefix+"maximize", "Maximize", "", respond.Action, "", nil)
	}

	for _, m := range monitorList {
		wd.Links = wd.Links.Add(actionPrefix+"move&monitor="+m.Name, "Move to monitor "+m.Name, "", respond.Action, "", nil)
	}

	return &wd

}

func (win Window) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, win.ToData())
	} else if r.Method == "POST" {
		var action = requests.Action(r)
		switch action {
		case "":
			dataMutex.Lock()
			defer dataMutex.Unlock()
			dataConnection.RaiseAndFocusWindow(uint32(win))
			respond.Accepted(w)
		case "restore":
			dataMutex.Lock()
			defer dataMutex.Unlock()
			if states, err := dataConnection.GetState(uint32(win)); err != nil {
				respond.NotFound(w)
			} else if slice.Contains(states, NET_WM_STATE_HIDDEN) {
				dataConnection.RemoveWmState(uint32(win), NET_WM_STATE_HIDDEN)
			} else {
				dataConnection.RemoveWmState(uint32(win), NET_WM_STATE_MAXIMIZED_HORZ)
				dataConnection.RemoveWmState(uint32(win), NET_WM_STATE_MAXIMIZED_VERT)
			}
		case "maximize":
			fmt.Println("maximizing")
			dataMutex.Lock()
			defer dataMutex.Unlock()
			dataConnection.AddWmState(uint32(win), NET_WM_STATE_MAXIMIZED_HORZ)
			dataConnection.AddWmState(uint32(win), NET_WM_STATE_MAXIMIZED_VERT)
			respond.Accepted(w)
		case "minimize":
			dataMutex.Lock()
			defer dataMutex.Unlock()
			dataConnection.AddWmState(uint32(win), NET_WM_STATE_HIDDEN)
			respond.Accepted(w)
		case "move":
			monitorName := requests.GetSingleQueryParameter(r, "monitor", "")
			fmt.Println("moving to ", monitorName)
			dataMutex.Lock()
			defer dataMutex.Unlock()
			if err := dataConnection.MoveToMonitor(uint32(win), monitorName); err != nil {
				respond.UnprocessableEntity(w, err)
			} else {
				respond.Accepted(w)
			}
		default:
			respond.UnprocessableEntity(w, fmt.Errorf("Unknown action %s", action))
		}
	} else if r.Method == "DELETE" {
		dataMutex.Lock()
		defer dataMutex.Unlock()
		dataConnection.CloseWindow(uint32(win))
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
	dataMutex.Lock()
	defer dataMutex.Unlock()

	return dataConnection.GetScreenshotAsPng(wId, downscale)
}

var dataConnection = MakeConnection()
var dataMutex = &sync.Mutex{}

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

func GetIconName(wId uint32) (string, error) {
	if name, ok := getIconNameFromCache(wId); ok {
		return name, nil
	} else {
		pixelArray, err := dataConnection.GetIcon(wId)
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
