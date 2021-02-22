// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"fmt"
	"sync"
	"time"

	"github.com/surlykke/RefudeServices/watch"
)

// Maintains windows  and monitors lists
func Run() {
	var c = MakeDisplay()
	var windows, _ = updateWindowList(c, []*Window{}, []*MonitorData{})
	repo.Lock()
	repo.desktoplayout, repo.windows = getDesktopLayout(c, windows)
	repo.Unlock()

	SubscribeToEvents(c)

	for {
		event, wId := fetchX11Events(c)
		var change = false
		repo.Lock()
		switch event {
		case DesktopStacking:
			repo.windows, change = updateWindowList(c, repo.windows, repo.desktoplayout.Monitors)
		case DesktopGeometry:
			repo.desktoplayout, repo.windows = getDesktopLayout(c, repo.windows)
			change = true
		case WindowTitle:
			repo.windows, change = updateWindow(repo.windows, wId, repo.desktoplayout.Monitors, func(w *Window) { w.Name, _ = GetName(c, w.Id) })
		case WindowIconName:
			repo.windows, change = updateWindow(repo.windows, wId, repo.desktoplayout.Monitors, func(w *Window) { w.IconName, _ = GetIconName(c, w.Id) })
		case WindowSt:
			repo.windows, change = updateWindow(repo.windows, wId, repo.desktoplayout.Monitors, func(w *Window) { w.State, _ = GetState(c, w.Id) })
		case WindowGeometry:
			updateWindowGeometry(c, wId)
		}
		change = change && repo.savedWindowsList == nil
		repo.Unlock()
		if change {
			watch.DesktopSearchMayHaveChanged()
		}
	}
}

type Repo struct {
	sync.Mutex
	windows             []*Window
	desktoplayout       *DesktopLayout
	highlightedWindowId uint32
	highlightDeadline   time.Time
	savedWindowsList    []*Window
}

func (r *Repo) windowsForServing() []*Window {
	r.Lock()
	defer r.Unlock()
	if r.savedWindowsList != nil {
		return r.savedWindowsList
	} else {
		return r.windows
	}
}

var repo = &Repo{windows: []*Window{}, desktoplayout: &DesktopLayout{}}

type Event uint8

const (
	DesktopGeometry Event = iota
	DesktopStacking
	WindowTitle
	WindowIconName
	WindowGeometry
	WindowSt
)

var EventNames = map[Event]string{
	DesktopGeometry: "DesktopGeometry",
	DesktopStacking: "DesktopStacking",
	WindowTitle:     "WindowTitle",
	WindowIconName:  "WindowIconName",
	WindowGeometry:  "WindowGeometry",
	WindowSt:        "WindowSt",
}

func (e Event) String() string {
	return EventNames[e]
}

func getMonitorList(c *Connection, windows []*Window) ([]*MonitorData, []*Window) {
	fmt.Println("updateMonitorList")
	var newMonitorList = c.GetMonitorDataList()

	// Update links on all windows
	var newWindowList = make([]*Window, len(repo.windows), len(repo.windows))
	for i, window := range repo.windows {
		var copy = *window
		updateLinksSingle(&copy, newMonitorList)
		newWindowList[i] = &copy
	}

	return newMonitorList, newWindowList
}

func updateWindowList(c *Connection, oldWindowList []*Window, monitors []*MonitorData) ([]*Window, bool) {
	var wIds = GetStack(c)
	var newWindowList = make([]*Window, len(wIds), len(wIds))
outerloop:
	for i, wId := range wIds {
		for _, oldWin := range oldWindowList {
			if oldWin.Id == wId {
				newWindowList[i] = oldWin
				continue outerloop
			}
		}
		newWindow := BuildWindow(c, wId)
		updateLinksSingle(newWindow, monitors)
		newWindowList[i] = newWindow
		SubscribeToWindowEvents(c, wId)
	}

	// Unless we have the same windows (excluding windows that do not show in desktopsearch) and in same order
	// desktopsearch may have changed
	var somethingChanged bool
	var i, j = 0, 0
	for {
		// skip what's not shown in desktopsearch
		for i < len(oldWindowList) && (oldWindowList[i].State&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) > 0) {
			i++
		}
		for j < len(newWindowList) && (newWindowList[j].State&(SKIP_TASKBAR|SKIP_PAGER|ABOVE) > 0) {
			j++
		}

		// Check
		if i >= len(oldWindowList) && j >= len(newWindowList) {
			break
		} else if i >= len(oldWindowList) || j >= len(newWindowList) || oldWindowList[i].Id != newWindowList[j].Id {
			somethingChanged = true
			break
		} else {
			i++
			j++
		}
	}

	return newWindowList, somethingChanged
}

func findWindow(windows []*Window, wId uint32) *Window {
	for _, win := range windows {
		if win.Id == wId {
			return win
		}
	}
	return nil
}

func findMonitor(monitors []*MonitorData, name string) *MonitorData {
	for _, m := range monitors {
		if m.Name == name {
			return m
		}
	}
	return nil
}

func replaceWindow(windows []*Window, newWin *Window) []*Window {
	var newWindowList = make([]*Window, len(windows), len(windows))
	for i, win := range windows {
		if win.Id == newWin.Id {
			newWindowList[i] = newWin
		} else {
			newWindowList[i] = win
		}
	}
	return newWindowList
}

func updateWindow(windows []*Window, wId uint32, monitors []*MonitorData, updater func(*Window)) ([]*Window, bool) {
	var newList = make([]*Window, len(windows), len(windows))
	var found = false
	for i, win := range windows {
		if win.Id == wId {
			var copy = *win
			updater(&copy)
			updateLinksSingle(&copy, monitors)
			newList[i] = &copy
			found = true
		} else {
			newList[i] = win
		}
	}
	return newList, found
}

const highlightTimeout = 3 * time.Second
const OPACITY uint32 = 0x11111111

func highlighWindow(wId uint32) {
	fmt.Println("Attempt to highligt", wId)
	if repo.savedWindowsList == nil {
		repo.savedWindowsList = repo.windows
		repo.highlightDeadline = time.Now().Add(highlightTimeout)
		scheduleUnhighlight(repo.highlightDeadline)
		for _, win := range repo.savedWindowsList {
			if win.Id != wId && win.State&(HIDDEN|ABOVE) == 0 {
				dataConnection.SetTransparent(win.Id, OPACITY)
			}
		}

	} else {
		repo.highlightDeadline = time.Now().Add(highlightTimeout)
		dataConnection.SetTransparent(repo.highlightedWindowId, OPACITY)
	}
	dataConnection.SetOpaque(wId)
	RaiseWindow(dataConnection, wId)
	repo.highlightedWindowId = wId
	fmt.Println("Highlighting done")
}

// Caller takes lock
func unHighligt() {
	for i := len(repo.savedWindowsList) - 1; i >= 0; i-- {
		dataConnection.SetOpaque(repo.savedWindowsList[i].Id)
		RaiseWindow(dataConnection, repo.savedWindowsList[i].Id)
	}
	repo.savedWindowsList = nil
}

func scheduleUnhighlight(at time.Time) {
	time.AfterFunc(at.Sub(time.Now())+100*time.Millisecond, func() {
		repo.Lock()
		defer repo.Unlock()
		if repo.highlightDeadline.After(time.Now()) {
			scheduleUnhighlight(repo.highlightDeadline)
		} else {
			unHighligt()
		}
	})
}

func updateWindowGeometry(c *Connection, wId uint32) {
	// TODO
}
