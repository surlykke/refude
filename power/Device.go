// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/lib/link"
)

type Device struct {
	DbusPath         dbus.ObjectPath
	title            string
	NativePath       string
	Vendor           string
	Model            string
	Serial           string
	UpdateTime       uint64
	Type             string
	IconName         string
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
	Percentage       int8
	IsPresent        bool
	State            string
	IsRechargeable   bool
	Capacity         float64
	Technology       string
	DisplayDevice    bool
}

func deviceTitle(devType, model string) string {
	// Try to, with the info we have from UPower, make a meaningful Title and Comment
	switch devType {
	case "Unknown":
		return "Unknown power device"
	case "Line Power":
		return "Line Power"
	case "Battery":
		return "Battery " + model
	default:
		return model
	}
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

func devicePath(path dbus.ObjectPath) string {
	if path == displayDeviceDbusPath {
		return displayDevicePath
	} else {
		return "/device" + string(path)
	}
}

func (d *Device) Self() string {
	return devicePath(d.DbusPath)	
}

func (d *Device) Presentation() (title string, comment string, icon link.Href, profile string) {
	return d.title, "", link.IconUrl(d.IconName), "device"
}

func (d *Device) Links(term string) link.List {
	return link.List{}
}

