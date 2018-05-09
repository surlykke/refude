// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"github.com/godbus/dbus"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

const PowerActionMediaType mediatype.MediaType = "application/vnd.org.refude.poweraction+json"

type PowerAction struct {
	Id            string
	Name          string
	Comment       string
	IconName      string
	Self          string
}

func (pa *PowerAction) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling: ", login1Service, ", ", login1Path, ", ", managerInterface + "." + pa.Id)
	dbusConn.Object(login1Service, login1Path).Call(managerInterface + "." + pa.Id, dbus.Flags(0), false)
	w.WriteHeader(http.StatusAccepted)
}
