// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package utils

import (
	"fmt"

	"github.com/godbus/dbus/v5"
)

const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"

func GetSingleProp(conn *dbus.Conn, service string, path dbus.ObjectPath, dbusInterface string, propName string) (dbus.Variant, bool) {
	if call := conn.Object(service, path).Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), dbusInterface, propName); call.Err != nil {
		return dbus.Variant{}, false
	} else {
		return call.Body[0].(dbus.Variant), true
	}
}

func GetAllProps(conn *dbus.Conn, service string, dbusPath dbus.ObjectPath, dbusInterface string) map[string]dbus.Variant {
	if call := conn.Object(service, dbusPath).Call(PROPERTIES_INTERFACE+".GetAll", dbus.Flags(0), dbusInterface); call.Err != nil {
		return map[string]dbus.Variant{}
	} else {
		return call.Body[0].(map[string]dbus.Variant)
	}
}

func Introspect(conn *dbus.Conn, service string, dbusPath dbus.ObjectPath) string {
	if call := conn.Object(service, dbusPath).Call(INTROSPECT_INTERFACE+".Introspect", dbus.Flags(0)); call.Err != nil {
		return fmt.Sprintln(call.Err)
	} else {
		return call.Body[0].(string)
	}
}
