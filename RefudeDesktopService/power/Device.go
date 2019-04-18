// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package power

import (
	"strings"
	"sync"

	"github.com/godbus/dbus"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const DeviceMediaType resource.MediaType = "application/vnd.org.refude.upowerdevice+json"

const SessionMediaType resource.MediaType = "application/vnd.org.refude.session+json"

var devices = make(map[resource.StandardizedPath]*Device)
var lock sync.Mutex

func GetDevice(path resource.StandardizedPath) *Device {
	lock.Lock()
	defer lock.Unlock()

	return devices[path]
}

func setDevice(device *Device) {
	lock.Lock()
	defer lock.Unlock()

	devices[device.GetSelf()] = device
}

func GetDevices() []interface{} {
	lock.Lock()
	defer lock.Unlock()

	var result = make([]interface{}, 0, len(devices))
	for _, device := range devices {
		result = append(result, device)
	}

	return result
}

var Session = buildSessionResource()

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
