// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	resource.Actions
	wId uint32
}

func MakeWindow(wId uint32) *Window {
	var w = &Window{
		resource.Links{windowSelf(wId), "window"},
		resource.Actions{},
		wId,
	}
	w.SetPostAction("default", resource.ResourceAction{Description: "Raise and focus", Executer: makeExecuter(w.wId)})
	return w
}

type ScreenShot uint32

func (ss ScreenShot) GetSelf() string {
	return fmt.Sprintf("/window/%d/screenshot", ss)
}

func (ss ScreenShot) GET(w http.ResponseWriter, r *http.Request) {
	var downscaleS = requests.GetSingleQueryParameter(r, "downscale", "1")
	var downscale = downscaleS[0] - '0'
	if downscale < 1 || downscale > 5 {
		requests.ReportUnprocessableEntity(w, fmt.Errorf("downscale should be >= 1 and <= 5"))
	} else if bytes, err := getScreenshot(uint32(ss), downscale); err == nil {
		w.Header().Set("Content-Type", "image/png")
		w.Write(bytes)
	} else {
		fmt.Println("Error getting screenshot:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

var dataConnection = xlib.MakeConnection()
var dataMutex sync.Mutex

func (w *Window) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	var data, err = json.Marshal(w.Links)
	if err != nil {
		return nil, err
	}
	buf.Write(data[:len(data)-1]) // omit last '}'
	fmt.Fprintf(&buf, `,"Id":%d`, w.wId)
	dataConnection.Lock()
	defer dataConnection.Unlock()
	if parent, err := dataConnection.GetParent(w.wId); err == nil {
		var X, Y int32
		var H, W uint32
		if parent != 0 {
			X, Y, W, H, err = dataConnection.GetGeometry(parent)
		} else {
			X, Y, W, H, err = dataConnection.GetGeometry(w.wId)
		}
		if err == nil {
			fmt.Fprintf(&buf, `,"X":%d,"Y":%d,"W":%d,"H":%d`, X, Y, W, H)
		}
	}
	if name, err := dataConnection.GetName(w.wId); err == nil {
		fmt.Fprintf(&buf, `,"Name": "%s"`, name)
	}
	if iconName, err := GetIconName(w.wId); err == nil {
		fmt.Fprintf(&buf, `,"IconName":"%s"`, iconName)
	}
	if states, err := dataConnection.GetState(w.wId); err == nil {
		fmt.Fprint(&buf, `,"States":[`)
		if len(states) > 0 {
			fmt.Fprintf(&buf, `"%s"`, states[0])
			for i := 1; i < len(states); i++ {
				fmt.Fprintf(&buf, `,"%s"`, states[i])
			}
		}
	}
	fmt.Fprint(&buf, `]}`)

	return buf.Bytes(), nil
}

func makeExecuter(wId uint32) func() {
	return func() {
		dataConnection.Lock()
		defer dataConnection.Unlock()
		dataConnection.RaiseAndFocusWindow(wId)
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
	dataConnection.Lock()
	defer dataConnection.Unlock()

	return dataConnection.GetScreenshotAsPng(wId, downscale)
}
