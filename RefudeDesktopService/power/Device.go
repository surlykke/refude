// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"io"
	"strings"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/serialize"
)

const DeviceMediaType resource.MediaType = "application/vnd.org.refude.upowerdevice+json"

const SessionMediaType resource.MediaType = "application/vnd.org.refude.session+json"

type Device struct {
	resource.GenericResource
	DbusPath         dbus.ObjectPath
	NativePath       string
	Vendor           string
	Model            string
	Serial           string
	UpdateTime       uint64
	Type             string
	PowerSupply      bool
	HasHistory       bool
	HasStatistics    bool
	Online           bool
	Energy           float64
	EnergyEmpty      float64
	EnergyFull       float64
	EnergyFullDesign float64
	EnergyRate       float64
	Voltage          float64
	TimeToEmpty      int64
	TimeToFull       int64
	Percentage       float64
	IsPresent        bool
	State            string
	IsRechargeable   bool
	Capacity         float64
	Technology       string
	DisplayDevice    bool
}

func deviceType(index uint32) string {
	var devType = []string{"Unknown", "Line Power", "Battery", "Ups", "Monitor", "Mouse", "Keyboard", "Pda", "Phone"}
	if index < 0 || index > 8 {
		index = 0
	}
	return devType[index]
}

func deviceState(index uint32) string {
	var devState = []string{"Unknown", "Charging", "Discharging", "Empty", "Fully charged", "Pending charge", "Pending discharge"}
	if index < 0 || index > 6 {
		index = 0
	}
	return devState[index]
}

func deviceTecnology(index uint32) string {
	var devTecnology = []string{"Unknown", "Lithium ion", "Lithium polymer", "Lithium iron phosphate", "Lead acid", "Nickel cadmium", "Nickel metal hydride"}
	if index < 0 || index > 6 {
		index = 0
	}
	return devTecnology[index]
}

func deviceSelf(dbusPath dbus.ObjectPath) resource.StandardizedPath {
	if strings.HasPrefix(string(dbusPath), DevicePrefix) {
		dbusPath = dbusPath[len(DevicePrefix):]
	}
	return resource.Standardizef("/device/%s", dbusPath)
}

func (d *Device) WriteBytes(w io.Writer) {
	d.GenericResource.WriteBytes(w)
	serialize.String(w, string(d.DbusPath))
	serialize.String(w, d.NativePath)
	serialize.String(w, d.Vendor)
	serialize.String(w, d.Model)
	serialize.String(w, d.Serial)
	serialize.UInt64(w, d.UpdateTime)
	serialize.String(w, d.Type)
	serialize.Bool(w, d.PowerSupply)
	serialize.Bool(w, d.HasHistory)
	serialize.Bool(w, d.HasStatistics)
	serialize.Bool(w, d.Online)
	serialize.Float64(w, d.Energy)
	serialize.Float64(w, d.EnergyEmpty)
	serialize.Float64(w, d.EnergyFull)
	serialize.Float64(w, d.EnergyFullDesign)
	serialize.Float64(w, d.EnergyRate)
	serialize.Float64(w, d.Voltage)
	serialize.UInt64(w, uint64(d.TimeToEmpty))
	serialize.UInt64(w, uint64(d.TimeToFull))
	serialize.Float64(w, d.Percentage)
	serialize.Bool(w, d.IsPresent)
	serialize.String(w, d.State)
	serialize.Bool(w, d.IsRechargeable)
	serialize.Float64(w, d.Capacity)
	serialize.String(w, d.Technology)
	serialize.Bool(w, d.DisplayDevice)
}
