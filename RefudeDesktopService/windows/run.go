// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package windows

import (
	"sync/atomic"

	"github.com/surlykke/RefudeServices/lib/respond"
)

// Maintains windows  and monitors lists
func Run() {
	var c = MakeConnection()
	c.SubscribeToEvents()
	storeWindowList(c)
	storeMonitorList(c)

	for {
		var i = WaitForEvent(c)
		if i == 1 {
			storeMonitorList(c)
		} else if i == 2 {
			storeWindowList(c)
		}
	}
}

var windows atomic.Value
var monitors atomic.Value

func init() {
	windows.Store([]uint32{})
	monitors.Store([]*Monitor{})
}

func storeMonitorList(c *Connection) {
	var monitorList = GetMonitors(c)
	for _, m := range monitorList {
		m.Links = respond.Links{{
			Href:  "/monitor/" + m.Name,
			Rel:   respond.Self,
			Title: m.Name,
		}}
	}
	monitors.Store(monitorList)
}

func storeWindowList(c *Connection) {
	windows.Store(GetStack(c))
}
