// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktopactions

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/tr"
)

var lastUpdated = atomic.Pointer[time.Time]{}

type StartResource struct {
	resource.ResourceData
}

var Start StartResource

func Run() {
	Start = StartResource{ResourceData: *resource.MakeBase("/start", "Refude desktop", "", "", mediatype.Start)}
	Start.AddLink("/start?action=shutdown", tr.Tr("Power off"), "/icon?name=system-shutdown", relation.Action)
	Start.AddLink("/start?action=reboot", tr.Tr("Reboot"), "/icon?name=system-reboot", relation.Action)
	Start.AddLink("/start?action=suspend", tr.Tr("Suspend"), "/icon?name=system-suspend", relation.Action)

	repo.Put(&Start)
}

func GetLinks(searchTerm string) []resource.Link {
	return Start.GetLinks()
}

func (s StartResource) DoPost(w http.ResponseWriter, r *http.Request) {
	var action = requests.GetSingleQueryParameter(r, "action", "-")
	var method string
	switch action {
	case "shutdown":
		method = "org.freedesktop.login1.Manager.PowerOff"
	case "reboot":
		method = "org.freedesktop.login1.Manager.Reboot"
	case "suspend":
		method = "org.freedesktop.login1.Manager.Suspend"
	default:
		respond.NotFound(w)
		return
	}

	if conn, err := dbus.SystemBus(); err != nil {
		respond.ServerError(w, err)
	} else {
		conn.Object("org.freedesktop.login1", "/org/freedesktop/login1").Call(method, dbus.Flags(0), false)
		respond.Accepted(w)
	}
}
