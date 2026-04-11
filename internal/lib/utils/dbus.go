// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package utils

import (
	"github.com/godbus/dbus/v5"
)

const PROPERTIES_INTERFACE = "org.freedesktop.DBus.Properties"

func Call[T any](conn *dbus.Conn, service string, dest dbus.ObjectPath, method string, args ...any) (T, error) {
	var t T
	var obj = conn.Object(service, dest)
	var call = obj.Call(method, 0, args...)
	if call.Err != nil {
		return t, call.Err
	} else if err := call.Store(&t); err != nil {
		return t, err
	} else {
		return t, nil
	}
}

func Prop[T any](conn *dbus.Conn, service string, dest dbus.ObjectPath, iface string, name string) (T, error) {
	return Call[T](conn, service, dest, "org.freedesktop.DBus.Properties.Get", iface, name)
}

func Props(conn *dbus.Conn, service string, dest dbus.ObjectPath, iface string) (map[string]any, error) {
	return Call[map[string]any](conn, service, dest, "org.freedesktop.DBus.Properties.GetAll", iface)
}
