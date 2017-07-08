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
	CanSuspend    bool
	CanHibernate  bool
	OnBattery     bool
	OnLowBattery  bool
	LidIsClosed   bool
	LidIsPresent  bool
}

type PropertyObject interface {
	ReadDBusProps(m map[string]dbus.Variant)
}

func (up *UPower) ReadDBusProps(m map[string]dbus.Variant) {
	for key, variant := range m {
		switch key {
		case "DaemonVersion":
			up.DaemonVersion = variant.Value().(string)
		case "CanSuspend":
			up.CanSuspend = variant.Value().(bool)
		case "CanHibernate":
			up.CanHibernate = variant.Value().(bool)
		case "OnBattery":
			up.OnBattery = variant.Value().(bool)
		case "OnLowBattery":
			up.OnLowBattery = variant.Value().(bool)
		case "LidIsClosed":
			up.LidIsClosed = variant.Value().(bool)
		case "LidIsPresent":
			up.LidIsPresent = variant.Value().(bool)
		}
	}
}

type Device struct {
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
}

func (d *Device) ReadDBusProps(m map[string]dbus.Variant) {
	for key, variant := range m {
		switch key {
		case "NativePath":
			d.NativePath = variant.Value().(string)
		case "Vendor":
			d.Vendor = variant.Value().(string)
		case "Model":
			d.Model = variant.Value().(string)
		case "Serial":
			d.Serial = variant.Value().(string)
		case "UpdateTime":
			d.UpdateTime = variant.Value().(uint64)
		case "Type":
			d.Type = deviceType(variant.Value().(uint32))
		case "PowerSupply":
			d.PowerSupply = variant.Value().(bool)
		case "HasHistory":
			d.HasHistory = variant.Value().(bool)
		case "HasStatistics":
			d.HasStatistics = variant.Value().(bool)
		case "Online":
			d.Online = variant.Value().(bool)
		case "Energy":
			d.Energy = variant.Value().(float64)
		case "EnergyEmpty":
			d.EnergyEmpty = variant.Value().(float64)
		case "EnergyFull":
			d.EnergyFull = variant.Value().(float64)
		case "EnergyFullDesign":
			d.EnergyFullDesign = variant.Value().(float64)
		case "EnergyRate":
			d.EnergyRate = variant.Value().(float64)
		case "Voltage":
			d.Voltage = variant.Value().(float64)
		case "TimeToEmpty":
			d.TimeToEmpty = variant.Value().(int64)
		case "TimeToFull":
			d.TimeToFull = variant.Value().(int64)
		case "Percentage":
			d.Percentage = variant.Value().(float64)
		case "IsPresent":
			d.IsPresent = variant.Value().(bool)
		case "State":
			d.State = deviceState(variant.Value().(uint32))
		case "IsRechargeable":
			d.IsRechargeable = variant.Value().(bool)
		case "Capacity":
			d.Capacity = variant.Value().(float64)
		case "Technology":
			d.Technology = deviceTecnology(variant.Value().(uint32))
		}
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
	var devTecnology = []string{"Unknown", "Lithium ion", "Lithium polymer", "Lithium iron phosphate", "Lead acid", "Nickel cadmium", "Nickel metal hydride" }
	if index < 0 || index > 6 {
		index = 0
	}
	return devTecnology[index]
}

