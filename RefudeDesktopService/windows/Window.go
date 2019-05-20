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
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows/xlib"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type Window struct {
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

type WindowCollection struct{}

type WinDmpResource uint32

func (wdr WinDmpResource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if bytes, err := xlib.GetScreenshotAsPng(uint32(wdr)); err == nil {
		w.Header().Set("Content-Type", "image/png")
		w.Write(bytes)
	} else {
		fmt.Println("Error getting screenshot:", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (wdr WinDmpResource) GetEtag() string {
	return ""
}

func (wc WindowCollection) Get(path string) resource.Res {
	if path == "/windows/brief" {
		if wIds, err := xlib.GetStack(); err != nil {
			return nil
		} else {
			var paths = make([]string, 0, len(wIds))
			for _, wId := range wIds {
				paths = append(paths, windowSelf(wId))
			}
			return resource.MakeJsonResource(paths)
		}
	} else if path == "/windows" {
		if windows, err := getWindows(); err != nil {
			return nil
		} else {
			return resource.MakeJsonResource(windows)
		}
	} else if strings.HasPrefix(path, "/window/") {
		if id, err := strconv.ParseUint(string(path[len("/window/"):]), 10, 32); err != nil {
			return nil
		} else {
			window, err := getWindow(uint32(id))
			if err != nil {
				return nil
			}
			return resource.MakeJsonResource(window)
		}
	} else if strings.HasPrefix(path, "/windmp/") {
		if id, err := strconv.ParseUint(string(path[len("/windmp/"):]), 10, 32); err != nil {
			return nil
		} else {
			return WinDmpResource(id)
		}
	} else {
		return nil
	}
}

func (wc WindowCollection) LongGet(path string, etagList string) resource.Res {
	return wc.Get(path)
}

var Windows = resource.MakeServer(WindowCollection{})
