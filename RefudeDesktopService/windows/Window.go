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
	"strconv"
	"strings"
	"sync"
)

const WindowMediaType resource.MediaType = "application/vnd.org.refude.wmwindow+json"

type Window struct {
	resource.AbstractResource
	Id            uint32
	Parent        uint32
	StackOrder    int
	X,Y,W,H       int
	Name          string
	IconName      string `json:",omitempty"`
	States        []string
}

type WindowCollection struct {
	sync.Mutex
	server.JsonResponseCache
	windows map[uint32]*Window
}

func MakeWindowCollection() *WindowCollection {
	var wc = &WindowCollection{}
	wc.JsonResponseCache = server.MakeJsonResponseCache(wc)
	wc.windows = make(map[uint32]*Window)
	return wc
}

func (wc *WindowCollection) getCopy(windowId uint32) *Window {
	if window, ok := wc.windows[windowId]; ok {
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

func (dac *WindowCollection) GetResource(r *http.Request) (interface{}, error) {
	var path = r.URL.Path
	if path == "/windows" {
		var windows = make([]*Window, 0, len(dac.windows))

		var matcher, err = requests.GetMatcher(r);
		if err != nil {
			return nil, err
		}

		for _, window := range dac.windows {
			if matcher(window) {
				windows = append(windows, window)
			}
		}

		return windows, nil
	} else if strings.HasPrefix(path, "/window/") {
		if id, err := strconv.ParseUint(path[len("/window/"):], 10, 32); err != nil {
			return nil, nil
		} else if window, ok := dac.windows[uint32(id)]; ok {
			return window, nil
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}

}


