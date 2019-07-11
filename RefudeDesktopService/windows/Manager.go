// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"log"
	"sync"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
)

// Returns windows in descending stack order
func getWindows() ([]interface{}, error) {
	if wIds, err := xlib.GetStack(); err != nil {
		return nil, fmt.Errorf("Unable to get client list stacking %v", err)
	} else {
		var windows = make([]interface{}, 0, len(wIds))
		for i := 0; i < len(wIds); i++ {
			var wId = wIds[len(wIds)-i-1]
			if window, err := getWindow(wId); err != nil {
				log.Printf("Error getting window %d: %v\n", wId, err)
			} else {
				windows = append(windows, window)
			}
		}

		return windows, nil
	}
}

func getWindow(wId uint32) (*Window, error) {
	window := &Window{}
	window.Id = wId
	window.Init(windowSelf(wId), "window")
	var err error
	window.Parent, err = xlib.GetParent(wId)
	if err != nil {
		return nil, err
	}

	if window.Parent != 0 {
		window.X, window.Y, window.W, window.H, err = xlib.GetGeometry(window.Parent)
	} else {
		window.X, window.Y, window.W, window.H, err = xlib.GetGeometry(wId)
	}
	if err != nil {
		return nil, err
	}

	window.Name, err = xlib.GetName(wId)
	if err != nil {
		return nil, err
	}

	window.IconName, err = GetIconName(wId)
	if err != nil {
		return nil, err
	}

	window.States, err = xlib.GetState(wId)
	if err != nil {
		return nil, err
	}

	var wIdCopy = wId
	var executer = func() {
		xlib.RaiseAndFocusWindow(wIdCopy)
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

		pixelArray, err := xlib.GetIcon(wId)
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
		var iconName = icons.AddPngFromARGB(icon)
		setIconNameInCache(wId, iconName)
		return iconName, nil
	}
}
