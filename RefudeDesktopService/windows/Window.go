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

	"github.com/surlykke/RefudeServices/lib/slice"

	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib/image"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"
)

type WindowData struct {
	Id       uint32
	Parent   uint32
	X, Y     int32
	W, H     uint32
	Name     string
	IconName string `json:",omitempty"`
	States   []string
}

type Window struct {
	self string
	id   uint32
}

func (w Window) ToStandardFormat() *respond.StandardFormat {
	var wd = WindowData{Id: w.id}
	dataMutex.Lock()
	if parent, err := dataConnection.GetParent(w.id); err == nil {
		if parent != 0 {
			wd.X, wd.Y, wd.W, wd.H, err = dataConnection.GetGeometry(parent)
		} else {
			wd.X, wd.Y, wd.W, wd.H, err = dataConnection.GetGeometry(w.id)
		}
	}
	if name, err := dataConnection.GetName(w.id); err == nil {
		wd.Name = name
	}
	if iconName, err := GetIconName(w.id); err == nil {
		wd.IconName = iconName
	}
	if states, err := dataConnection.GetState(w.id); err == nil {
		wd.States = states
	}
	defer dataMutex.Unlock()

	return &respond.StandardFormat{
		Self:      fmt.Sprintf("/window/%d", w.id),
		Type:      "window",
		Title:     wd.Name,
		OnPost:    "Raise and focus",
		IconName:  wd.IconName,
		Data:      wd,
		NoDisplay: slice.Contains(wd.States, "_NET_WM_STATE_ABOVE", "_NET_WM_STATE_SKIP_TASKBAR"),
	}
}

var dataConnection = xlib.MakeConnection()
var dataMutex = &sync.Mutex{}

func makeExecuter(wId uint32) func(http.ResponseWriter, *http.Request) {
	return func(http.ResponseWriter, *http.Request) {
	}
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
	dataMutex.Lock()
	defer dataMutex.Unlock()

	return dataConnection.GetScreenshotAsPng(wId, downscale)
}
