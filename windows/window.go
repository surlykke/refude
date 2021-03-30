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
	"net/url"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type Bounds struct {
	X, Y int32
	W, H uint32
}

type Geometry uint32

func (g Geometry) MarshalJSON() ([]byte, error) {
	return respond.ToJson(GetBounds(uint32(g))), nil
}

type Window struct {
	respond.Resource
	Id       uint32
	Parent   uint32
	Name     string
	IconName string `json:",omitempty"`
	Geometry Geometry
	State    x11.WindowStateMask
	Stacking int // 0 means: on top, then 1, then 2 etc. -1 means we don't know (yet)
}

// Caller ensures thread safety
func BuildWindow(p x11.Proxy, wId uint32) *Window {
	var win = &Window{Id: wId}
	win.Parent, _ = x11.GetParent(p, wId)
	win.Name, _ = x11.GetName(p, wId)
	win.IconName, _ = GetIconName(p, wId)
	if win.Parent > 0 {
		win.Geometry = Geometry(win.Parent)
	} else {
		win.Geometry = Geometry(win.Id)
	}
	win.State = x11.GetStates(p, wId)
	win.Stacking = -1
	var href = fmt.Sprintf("/window/%d", win.Id)
	win.Resource = respond.MakeResource(href, win.Name, icons.IconUrl(win.IconName), win, "window")
	var closeAction = respond.MakeAction("", "Close window", "window-close", func(*http.Request) error {
		x11.CloseWindow(requestProxy, win.Id)
		return nil
	})
	win.Self.Options.DELETE = &closeAction
	return win
}

func updateLinks(win *Window, desktopLayout *DesktopLayout) {
	win.ClearActions()
	//win.Links = win.Links.Add(href+"/screenshot", "Screenshot of "+win.Name, "", respond.Related, "/profile/window-screenshot", nil)
	win.AddAction(respond.MakeAction("", "Raise and focus", "", MakeRaiser(win.Id)))

	if win.State.Is(x11.HIDDEN) || win.State.Is(x11.MAXIMIZED_HORZ|x11.MAXIMIZED_VERT) {
		win.AddAction(respond.MakeAction("restore", "Restore window", "", MakeRestorer(win.Id)))
	} else {
		win.AddAction(respond.MakeAction("minimize", "Minimize window", "", MakeMinimizer(win.Id)))
		win.AddAction(respond.MakeAction("maximize", "Maximize window", "", MakeMaximizer(win.Id)))
	}

	for _, m := range desktopLayout.Monitors {
		var actionId = url.QueryEscape("move::" + m.Name)
		win.AddAction(respond.MakeAction(actionId, "Move to monitor "+m.Name, "", MakeMover(win.Id, m.Name)))
	}

	win.AddAction(respond.MakeAction("highlight", "Highlight window", "", MakeHighlighter(win.Id)))
}

func (win *Window) copy() *Window {
	var result = &Window{}
	*result = *win
	result.Owner = result
	return result
}

func relevantForDesktopSearch(w *Window) bool {
	return w.State&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0
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
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	return x11.GetScreenshotAsPng(requestProxy, wId, downscale)
}

func GetIconName(p x11.Proxy, wId uint32) (string, error) {
	pixelArray, err := x11.GetIcon(p, wId)
	if err != nil {
		fmt.Println("Error converting x11 icon to pngs", err)
		return "", err
	} else {
		return icons.AddX11Icon(pixelArray)
	}
}

var noBounds = &Bounds{0, 0, 0, 0}

func GetBounds(wId uint32) *Bounds {
	requestProxyMutex.Lock()
	defer requestProxyMutex.Unlock()
	// TODO Perhaps some caching
	if x, y, w, h, err := x11.GetGeometry(requestProxy, wId); err != nil {
		return noBounds
	} else {
		return &Bounds{x, y, w, h}
	}
}

type WindowStack []*Window

// Implement sort.Interface
func (ws WindowStack) Len() int {
	return len(ws)
}

func (ws WindowStack) Less(i, j int) bool {
	return ws[i].Stacking < ws[j].Stacking
}

func (ws WindowStack) Swap(i, j int) {
	ws[i], ws[j] = ws[j], ws[i]
}
