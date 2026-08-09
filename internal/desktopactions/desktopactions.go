// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktopactions

import (
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/resource"
)

type StartResource struct {
	resource.Base
	dbusMethod string
}

func makeStartResource(path string, title string, icon string, dbusMethod string) *StartResource {
	var sr = &StartResource{Base: resource.MakeBase(path, title, "Desktop action", icon), dbusMethod: dbusMethod}
	sr.AddAction("", "", resource.OrgRefudeAction, "")
	return sr
}

func Run() {
	resource.Put(makeStartResource("/start/shutdown", "Power off", "system-shutdown", "org.freedesktop.login1.Manager.PowerOff"))
	resource.Put(makeStartResource("/start/reboot", "Reboot", "system-reboot", "org.freedesktop.login1.Manager.Reboot"))
	resource.Put(makeStartResource("/start/suspend", "Suspend", "system-suspend", "org.freedesktop.login1.Manager.Suspend"))

}

func (this *StartResource) DoPost(action string) (bool, error) {
	if conn, err := dbus.SystemBus(); err != nil {
		log.Print(err)
		return false, err
	} else {
		conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").Call(this.dbusMethod, dbus.Flags(0), false)
		return true, nil
	}

}
