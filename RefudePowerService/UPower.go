// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/godbus/dbus"
	"net/http"
)

type UPowerObject interface {
	http.Handler
	ReadDBusProps(m map[string]dbus.Variant)
	Copy() UPowerObject
}

type UPower struct {
	DaemonVersion string
	OnBattery     bool
	OnLowBattery  bool
	LidIsClosed   bool
	LidIsPresent  bool
}


