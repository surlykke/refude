// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"net/http"
	"sync"
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

type WindowCollection struct {
	mutex   sync.Mutex
	windows map[resource.StandardizedPath]*Window
	server.CachingJsonGetter
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func (*WindowCollection) HandledPrefixes() []string {
	return []string{"/window"}
}

func MakeWindowCollection() *WindowCollection {
	var wc = &WindowCollection{}
	wc.CachingJsonGetter = server.MakeCachingJsonGetter(wc)
	wc.windows = make(map[resource.StandardizedPath]*Window)
	return wc
}

func (wc *WindowCollection) GetSingle(r *http.Request) interface{} {
	wc.mutex.Lock()
	defer wc.mutex.Unlock()
	if window, ok := wc.windows[resource.Standardize(r.URL.Path)]; ok {
		return window
	} else {
		return nil
	}
}

func (wc *WindowCollection) GetCollection(r *http.Request) []interface{} {
	if r.URL.Path == "/windows" {
		var windows = make([]interface{}, 0, len(wc.windows))

		for _, window := range wc.windows {
			windows = append(windows, window)
		}

		return windows
	} else {
		return nil
	}
}

func (wc *WindowCollection) POST(w http.ResponseWriter, r *http.Request) {
	if res := wc.GetSingle(r); res == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if window, ok := res.(*Window); !ok {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var actionId = requests.GetSingleQueryParameter(r, "action", "default")
		if action, ok := window.ResourceActions[actionId]; ok {
			action.Executer()
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}
	}
}



func (wc *WindowCollection) getCopy(windowId uint32) *Window {
	if window, ok := wc.windows[windowSelf(windowId)]; ok {
		var copy = *window
		return &copy
	} else {
		return nil
	}

}

func (wc *WindowCollection) getCopyByParent(parent uint32) *Window {
	for _, window := range wc.windows {
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
