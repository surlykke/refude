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
	"sync/atomic"

	"github.com/surlykke/RefudeServices/watch"
)

// Maintains windows  and monitors lists
func Run() {
	var c = MakeDisplay()
	SubscribeToEvents(c)
	updateMonitorList(c)
	updateWindowList(c)

	for {
		if event, wId, err := NextEvent(c); err != nil {
			log.Println("Error from NextEvent", err)
		} else {
			switch event {
			case DesktopGeometry:
				updateMonitorList(c)
			case DesktopStacking:
				updateWindowList(c)
			case WindowTitle:
				updateWindowTitle(c, wId)
			case WindowIconName:
				updateWindowIconName(c, wId)
			case WindowSt:
				updateWindowState(c, wId)
			case WindowGeometry:
				updateWindowGeometry(c, wId)
			}
		}
	}
}

var deskopLayout atomic.Value
var windows atomic.Value
var monitors atomic.Value

func init() {
	deskopLayout.Store(&DesktopLayout{})
	windows.Store([]*Window{})
	monitors.Store([]*MonitorData{})
}

func updateMonitorList(c *Connection) {
	fmt.Println("updateMonitorList")
	var newMonitorList = c.GetMonitorDataList()
	var newDesktopLayout = BuildDesktopLayout(newMonitorList)
	monitors.Store(newMonitorList)
	deskopLayout.Store(newDesktopLayout)

	// Update links on all windows
	var windowList = windows.Load().([]*Window)
	var newWindowList = make([]*Window, len(windowList), len(windowList))
	for i, window := range windowList {
		var copy = *window
		BuildLinks(&copy)
		newWindowList[i] = &copy
	}
	windows.Store(newWindowList)

	updateWindowList(c)
	watch.SomethingChanged("/desktoplayout")
	watch.DesktopSearchMayHaveChanged()
}

func updateWindowList(c *Connection) {
	var wIds = GetStack(c)
	var windowList = windows.Load().([]*Window)
	var newWindowList = make([]*Window, len(wIds), len(wIds))
outerloop:
	for i, wId := range wIds {
		for _, oldWin := range windowList {
			if oldWin.Id == wId {
				newWindowList[i] = oldWin
				continue outerloop
			}
		}
		newWindowList[i] = BuildWindow(c, wId)
		SubscribeToWindowEvents(c, wId)
	}
	windows.Store(newWindowList)

	// Check that we have the same windows (excluding windows that do not show in desktopsearch) and in same order
	// Otherwise publish desktopsearch change
	var i, j = 0, 0
	for {
		// skip what's not shown in desktopsearch
		for i < len(windowList) && (windowList[i].State&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) > 0) {
			i++
		}
		for j < len(windowList) && (windowList[j].State&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) > 0) {
			j++
		}

		// Check
		if i >= len(windowList) && j >= len(newWindowList) {
			break
		} else if i >= len(windowList) || j >= len(newWindowList) || windowList[i].Id != newWindowList[j].Id {
			watch.DesktopSearchMayHaveChanged()
			break
		} else {
			i++
			j++
		}
	}
}

func findWindow(wId uint32) *Window {
	for _, win := range windows.Load().([]*Window) {
		if win.Id == wId {
			return win
		}
	}
	return nil
}

func replaceWindow(newWin *Window) {
	var windowList = windows.Load().([]*Window)
	var newWindowList = make([]*Window, len(windowList), len(windowList))
	for i, win := range windowList {
		if win.Id == newWin.Id {
			newWindowList[i] = newWin
		} else {
			newWindowList[i] = win
		}
	}
	windows.Store(newWindowList)
}

func updateWindowTitle(c *Connection, wId uint32) {
	if win := findWindow(wId); win != nil {
		var copy = *win
		copy.Name, _ = GetName(c, wId)
		BuildLinks(&copy)
		replaceWindow(&copy)
		watch.DesktopSearchMayHaveChanged()
	}
}

func updateWindowIconName(c *Connection, wId uint32) {
	if win := findWindow(wId); win != nil {
		var copy = *win
		copy.IconName, _ = GetIconName(c, wId)
		BuildLinks(&copy)
		replaceWindow(&copy)
		watch.DesktopSearchMayHaveChanged()
	}
}

func updateWindowState(c *Connection, wId uint32) {
	if win := findWindow(wId); win != nil {
		var copy = *win
		copy.State, _ = GetState(c, wId)
		BuildLinks(&copy)
		replaceWindow(&copy)
		watch.DesktopSearchMayHaveChanged()
	}
}

func updateWindowGeometry(c *Connection, wId uint32) {
	// TODO
}
