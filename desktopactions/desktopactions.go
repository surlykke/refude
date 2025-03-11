// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktopactions

import (
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/lib/tr"
)

var lastUpdated = atomic.Pointer[time.Time]{}

var actionMethod = map[string]string{
	"shutdown": "org.freedesktop.login1.Manager.PowerOff",
	"reboot":   "org.freedesktop.login1.Manager.Reboot",
	"suspend":  "org.freedesktop.login1.Manager.Suspend",
}

type StartResource struct {
	entity.Base
}

var Start StartResource

var Resources = []*StartResource{&Start}

func init() {
	Start = StartResource{Base: *entity.MakeBase("Refude desktop", "", mediatype.Start)}
	Start.Keywords = []string{"sluk", "genstart", "slumre"}
	Start.AddAction("shutdown", tr.Tr("Power off"), "system-shutdown")
	Start.AddAction("reboot", tr.Tr("Reboot"), "system-reboot")
	Start.AddAction("suspend", tr.Tr("Suspend"), "system-suspend")
	Start.SetPath("/start")
}

func GetHandler() response.Response {
	return response.Json(Start)
}

func PostHandler(action string) response.Response {
	if method, ok := actionMethod[action]; ok {
		if conn, err := dbus.SystemBus(); err != nil {
			return response.ServerError(err)
		} else {
			conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").Call(method, dbus.Flags(0), false)
			return response.Accepted()
		}
	} else {
		return response.NotFound()
	}
}
