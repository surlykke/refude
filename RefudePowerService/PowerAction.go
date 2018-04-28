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
	"github.com/surlykke/RefudeServices/lib/resource"
	"fmt"
)

const PowerActionMediaType resource.MediaType = "application/vnd.org.refude.poweraction+json"

type PowerAction struct {
	resource.ByteResource
	Id            string
	Name          string
	Comment       string
	IconName      string
	Can           bool
	RelevanceHint int
	Self          string
}

func NewPowerAction(Id string, Name string, Comment string, IconName string) *PowerAction {
	can := "yes" == dbusConn.Object(login1Service, login1Path).Call(managerInterface + ".Can" + Id, dbus.Flags(0)).Body[0].(string)
	return &PowerAction{resource.MakeByteResource(PowerActionMediaType), Id, Name, Comment, IconName, can, 0, ""}
}

func (pa *PowerAction) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Calling: ", login1Service, ", ", login1Path, ", ", managerInterface + "." + pa.Id)
	dbusConn.Object(login1Service, login1Path).Call(managerInterface + "." + pa.Id, dbus.Flags(0), false)
	w.WriteHeader(http.StatusAccepted)
}
