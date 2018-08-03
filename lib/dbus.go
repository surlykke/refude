// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import "github.com/godbus/dbus"

const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"

func GetSingleProp(conn *dbus.Conn, service string, path dbus.ObjectPath, dbusInterface string, propName string) (dbus.Variant, bool) {
	if call := conn.Object(service, path).Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), dbusInterface, propName); call.Err != nil {
		return dbus.Variant{}, false
	} else {
		return call.Body[0].(dbus.Variant), true
	}
}

func GetAllProps(conn *dbus.Conn, sender string, dbusPath dbus.ObjectPath, dbusInterface string) map[string]dbus.Variant {
	if call := conn.Object(sender, dbusPath).Call(PROPERTIES_INTERFACE+".GetAll", dbus.Flags(0), dbusInterface); call.Err != nil {
		return map[string]dbus.Variant{}
	} else {
		return call.Body[0].(map[string]dbus.Variant)
	}
}
