// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package power

import (
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/refude/internal/lib/entity"
)

type Device struct {
	entity.Base
	Id string

	// Property of our making
	DisplayDevice bool

	// Properties from upower/dbus
	// Albeit some of them translated to text
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
	Luminosity       float64
	TimeToEmpty      int64
	TimeToFull       int64
	Percentage       float64
	IsPresent        bool
	State            string
	IsRechargeable   bool
	Capacity         float64
	Technology       string
	Warninglevel     string
	Batterylevel     string
}

func deviceTitle(devType, model string) string {
	// Try to, with the info we have from UPower, make a meaningful Title and Comment
	switch devType {
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

func dbusPath2id(path dbus.ObjectPath) string {
	// I _think_ it's safe to assume all device paths start with devicePrefix
	if strings.HasPrefix(string(path), devicePrefix) {
		return strings.ReplaceAll(string(path)[len(devicePrefix):], "/", "-")
	} else if strings.HasPrefix(string(path), "/") {
		return string(path)[1:]
	} else {
		return string(path)
	}
}

func deviceWarningLevel(index uint32) string {
	var devWarningLevel = []string{"Unknown", "None", "Discharging", "Low", "Critical", "Action"}
	if index < 0 || index > 5 {
		index = 0
	}
	return devWarningLevel[index]
}

func deviceBatteryLevel(index uint32) string {
	var devBatteryLevel = []string{
		"Unknown",
		"None",
		"",
		"Low",
		"Critical",
		"",
		"Normal",
		"High",
		"Full",
	}
	if index < 0 || index > 8 || index == 2 || index == 5 {
		index = 0
	}
	return devBatteryLevel[index]
}
