// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/requests"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type WindowData struct {
	resource.Links
	Id         uint32
	Parent     uint32
	StackOrder int
	X, Y       int32
	W, H       uint32
	Name       string
	IconName   string `json:",omitempty"`
	States     []string
}

func windowSelf(windowId uint32) string {
	return fmt.Sprintf("/window/%d", windowId)
}

type Window struct {
	resource.Links
	wId uint32
}

func MakeWindow(wId uint32) *Window {
	var w = &Window{
		wId: wId,
	}
	w.Init(windowSelf(wId), "window")
	return w
}

type Windows struct {
	resource.Links
	wIds []uint32
}

func MakeWindows(wIds []uint32) *Windows {
	var windows = &Windows{
		wIds: wIds,
	}
	windows.Init("/windows", "windows")
	return windows
}

func (wdr Window) GET(w http.ResponseWriter, r *http.Request) {
	if requests.HaveParam(r, "screenshot") {
		var downscaleS = requests.GetSingleQueryParameter(r, "downscale", "1")
		var downscale = downscaleS[0] - '0'
		if downscale < 1 || downscale > 5 {
			requests.ReportUnprocessableEntity(w, fmt.Errorf("downscale should be >= 1 and <= 5"))
		} else if bytes, err := getScreenshot(wdr.wId, downscale); err == nil {
			w.Header().Set("Content-Type", "image/png")
			w.Write(bytes)
		} else {
			fmt.Println("Error getting screenshot:", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
	} else {
		window, err := getWindow(wdr.wId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			resource.ServeAsJson(w, r, window)
		}
	}
}

func (ws Windows) GET(w http.ResponseWriter, r *http.Request) {
	if requests.HaveParam(r, "brief") {
		var paths = make([]string, 0, len(ws.wIds))
		for _, wId := range ws.wIds {
			paths = append(paths, windowSelf(wId))
		}
		resource.ServeAsJson(w, r, paths)
	} else {
		var windows = make([]*WindowData, 0, len(ws.wIds))
		for _, wId := range ws.wIds {
			var window, err = getWindow(wId)
			if err == nil {
				windows = append(windows, window)
			} else {
				log.Println("WARN unable to retrieve window", wId)
			}
		}
		resource.ServeAsJson(w, r, windows)
	}
}

var dataConnection = xlib.MakeConnection()
var dataMutex sync.Mutex

func getWindow(wId uint32) (*WindowData, error) {
	dataConnection.Lock()
	defer dataConnection.Unlock()
	window := &WindowData{}
	window.Id = wId
	window.Init(windowSelf(wId), "window")
	var err error
	window.Parent, err = dataConnection.GetParent(wId)
	if err != nil {
		return nil, err
	}
	if window.Parent != 0 {
		window.X, window.Y, window.W, window.H, err = dataConnection.GetGeometry(window.Parent)
	} else {
		window.X, window.Y, window.W, window.H, err = dataConnection.GetGeometry(wId)
	}
	if err != nil {
		return nil, err
	}
	window.Name, err = dataConnection.GetName(wId)
	if err != nil {
		return nil, err
	}
	window.IconName, err = GetIconName(wId)
	if err != nil {
		return nil, err
	}
	window.States, err = dataConnection.GetState(wId)
	if err != nil {
		return nil, err
	}
	var wIdCopy = wId
	var executer = func() {
		dataConnection.Lock()
		defer dataConnection.Unlock()
		dataConnection.RaiseAndFocusWindow(wIdCopy)
	}

	window.AddAction("default", resource.ResourceAction{Description: "Raise and focus", IconName: window.IconName, Executer: executer})
	return window, nil
}

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

func getScreenshot(wId uint32, downscale byte) ([]byte, error) {
	dataConnection.Lock()
	defer dataConnection.Unlock()

	return dataConnection.GetScreenshotAsPng(wId, downscale)
}
