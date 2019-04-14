// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

const WindowMediaType resource.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.AbstractResource
	Id         uint32
	Parent     uint32
	StackOrder int
	X, Y, W, H int
	Name       string
	IconName   string `json:",omitempty"`
	States     []string
}

var windowCollection = make(map[resource.StandardizedPath]*Window)
var mutex sync.Mutex

func GetWindow(path resource.StandardizedPath) *Window {
	mutex.Lock()
	defer mutex.Unlock()

	return windowCollection[path]
}

func setWindow(win *Window) {
	mutex.Lock()
	defer mutex.Unlock()

	windowCollection[win.GetSelf()] = win
}

func GetWindows() []interface{} {
	mutex.Lock()
	defer mutex.Unlock()

	var windows = make([]interface{}, 0, len(windowCollection))
	for _, win := range windowCollection {
		windows = append(windows, win)
	}
	return windows
}

func ClearAll() {
	mutex.Lock()
	defer mutex.Unlock()

	windowCollection = make(map[resource.StandardizedPath]*Window)
}

func GetCopy(winId uint32) *Window {
	mutex.Lock()
	defer mutex.Unlock()

	for _, win := range windowCollection {
		if win.Id == winId {
			return &(*win)
		}
	}

	return nil
}

func getCopyByParent(parent uint32) *Window {
	mutex.Lock()
	defer mutex.Unlock()

	for _, window := range windowCollection {
		if window.Parent == parent {
			var copy = *window
			return &copy
		}
	}

	return nil
}

func windowSelf(windowId uint32) resource.StandardizedPath {
	return resource.Standardizef("/window/%d", windowId)
}
