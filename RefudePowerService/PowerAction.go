/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"net/http"
	"github.com/surlykke/RefudeServices/common"
	"github.com/godbus/dbus"
	"fmt"
)



type PowerAction struct {
	Id string
	Name string
	Comment string
	IconName string
	Can bool
}

func NewPowerAction(Id string, Name string, Comment string, IconName string) *PowerAction {
	can := "yes" == dbusConn.Object(login1Service, login1Path).Call(managerInterface + ".Can" + Id, dbus.Flags(0)).Body[0].(string)
	return &PowerAction{Id, Name, Comment, IconName, can}
}

func (p PowerAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println(p.Id, "POST, can: ", p.Can)
	if r.Method == "GET" {
		common.ServeAsJson(w, r, p)
	} else if r.Method == "POST" && p.Can {
		fmt.Println("Calling: ", login1Service, ", ", login1Path, ", ", managerInterface + "." + p.Id)
		dbusConn.Object(login1Service, login1Path).Call(managerInterface + "." + p.Id, dbus.Flags(0), false)
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

}


