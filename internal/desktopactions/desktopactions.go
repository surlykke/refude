// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktopactions

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/response"
)

var lastUpdated = atomic.Pointer[time.Time]{}

var actionMethod = map[string]string{
	"shutdown": "org.freedesktop.login1.Manager.PowerOff",
	"reboot":   "org.freedesktop.login1.Manager.Reboot",
	"suspend":  "org.freedesktop.login1.Manager.Suspend",
}

var PowerActions = entity.MakeMap[string, *StartResource]()

type StartResource struct {
	entity.Base
	dbusMethod string
}

func (this *StartResource) DoPost(action string) response.Response {
	if action != "" {
		return response.NotFound()
	} else if conn, err := dbus.SystemBus(); err != nil {
		log.Print(err)
		return response.ServerError(err)
	} else {
		conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").Call(this.dbusMethod, dbus.Flags(0), false)
		return response.Accepted()
	}

}

func init() {
	var datas = [][]string{
		{"shutdown", "Power off", "system-shutdown", "org.freedesktop.login1.Manager.PowerOff"},
		{"reboot", "Reboot", "system-reboot", "org.freedesktop.login1.Manager.Reboot"},
		{"suspend", "Suspend", "system-suspend", "org.freedesktop.login1.Manager.Suspend"}}

	for _, data := range datas {
		var res = StartResource{Base: *entity.MakeBase(data[1], "", data[2], "Power action"), dbusMethod: data[3]}
		res.AddAction("", "", "")
		PowerActions.Put(data[0], &res)
	}
}
