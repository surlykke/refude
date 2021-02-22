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

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Bounds struct {
	X, Y int32
	W, H uint32
}

type Geometry uint32

func (g Geometry) MarshalJSON() ([]byte, error) {
	return respond.ToJson(GetBounds(dataConnection, uint32(g))), nil
}

type Window struct {
	respond.Links `json:"_links"`
	Id            uint32
	Parent        uint32
	Name          string
	IconName      string `json:",omitempty"`
	Geometry      Geometry
	State         WindowStateMask
}

// Caller ensures thread safety
func BuildWindow(c *Connection, wId uint32) *Window {
	var win = &Window{Id: wId}
	win.Parent, _ = GetParent(c, wId)
	win.Name, _ = GetName(c, wId)
	win.IconName, _ = GetIconName(c, wId)
	if win.Parent > 0 {
		win.Geometry = Geometry(win.Parent)
	} else {
		win.Geometry = Geometry(win.Id)
	}
	win.State, _ = GetState(c, wId)

	return win
}

func updateLinksSingle(win *Window, monitorList []*MonitorData) {
	var href = fmt.Sprintf("/window/%d", win.Id)
	var actionPrefix = href + "?action="
	win.Links = make(respond.Links, 0, 5+len(monitorList))
	win.Links = win.Links.Add(href, win.Name, icons.IconUrl(win.IconName), respond.Self, "/profile/window", respond.Hints{"states": win.State})

	win.Links = win.Links.Add(href+"/screenshot", "Screenshot of "+win.Name, "", respond.Related, "/profile/window-screenshot", nil)
	//wd.Links = wd.Links.Add(actionPrefix+"raise", "Raise and focus", "", respond.Action, "", nil)

	if win.State&(HIDDEN|MAXIMIZED_HORZ|MAXIMIZED_VERT) != 0 {
		win.Links = win.Links.Add(actionPrefix+"restore", "Restore", "", respond.Action, "", nil)
	} else {
		win.Links = win.Links.Add(actionPrefix+"minimize", "Minimize", "", respond.Action, "", nil)
		win.Links = win.Links.Add(actionPrefix+"maximize", "Maximize", "", respond.Action, "", nil)
	}

	for _, m := range monitorList {
		win.Links = win.Links.Add(actionPrefix+"move&monitor="+m.Name, "Move to monitor "+m.Name, "", respond.Action, "", nil)
	}
}

func updateLinks(windowList []*Window, monitorList []*MonitorData, all bool) {
	for _, win := range windowList {
		if all || len(win.Links) == 0 {
			updateLinksSingle(win, monitorList)
		}
	}
}

func (win Window) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, win)
	} else if r.Method == "POST" {
		switch requests.Action(r) {
		case "":
			RaiseAndFocusWindow(dataConnection, win.Id)
			respond.Accepted(w)
		case "restore":
			RemoveStates(dataConnection, win.Id, HIDDEN|MAXIMIZED_HORZ|MAXIMIZED_VERT)
			respond.Accepted(w)
		case "maximize":
			RemoveStates(dataConnection, win.Id, HIDDEN)
			AddStates(dataConnection, win.Id, MAXIMIZED_HORZ|MAXIMIZED_VERT)
			respond.Accepted(w)
		case "minimize":
			AddStates(dataConnection, win.Id, HIDDEN)
			respond.Accepted(w)
		case "move":
			monitorName := requests.GetSingleQueryParameter(r, "monitor", "")
			repo.Lock()
			var monitors = repo.desktoplayout.Monitors
			repo.Unlock()
			if m := findMonitor(monitors, monitorName); m != nil {
				var saveStates = win.State & (HIDDEN | MAXIMIZED_HORZ | MAXIMIZED_VERT)
				RemoveStates(dataConnection, win.Id, HIDDEN|MAXIMIZED_HORZ|MAXIMIZED_VERT)
				var marginW, marginH = m.W / 10, m.H / 10
				SetBounds(dataConnection, win.Id, m.X+int32(marginW), m.Y+int32(marginH), m.W-2*marginW, m.H-2*marginH)
				AddStates(dataConnection, win.Id, saveStates)
			}
			respond.Accepted(w)
		case "highlight":
			repo.Lock()
			highlighWindow(win.Id)
			repo.Unlock()
			respond.Accepted(w)
		default:
			respond.UnprocessableEntity(w, fmt.Errorf("Unknown action '%s'", requests.Action(r)))
		}

	} else if r.Method == "DELETE" {
		dataConnection.CloseWindow(win.Id)
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
	return GetScreenshotAsPng(dataConnection, wId, downscale)
}

var dataConnection = MakeDisplay()

func GetIconName(c *Connection, wId uint32) (string, error) {
	pixelArray, err := GetIcon(c, wId)
	if err != nil {
		fmt.Println("Error converting x11 icon to pngs", err)
		return "", err
	} else {
		return icons.AddX11Icon(pixelArray)
	}
}

var noBounds = &Bounds{0, 0, 0, 0}

func GetBounds(c *Connection, wId uint32) *Bounds {
	// TODO Perhaps some caching
	if x, y, w, h, err := dataConnection.GetGeometry(wId); err != nil {
		return noBounds
	} else {
		return &Bounds{x, y, w, h}
	}
}
