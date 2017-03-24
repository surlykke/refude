/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"github.com/godbus/dbus"
	"net/http"
	"github.com/surlykke/RefudeServices/common"
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

func (up UPower) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	common.ServeGetAsJson(w, r, up)
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

func (d Device) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	common.ServeGetAsJson(w, r, d)
}

func deviceType(index uint32) string {
	switch index {
	case 0:
		return "Unknown"
	case 1:
		return "Line Power"
	case 2:
		return "Battery"
	case 3:
		return "Ups"
	case 4:
		return "Monitor"
	case 5:
		return "Mouse"
	case 6:
		return "Keyboard"
	case 7:
		return "Pda"
	case 8:
		return "Phone"
	default:
		return "Unknown"
	}
}

func deviceState(index uint32) string {
	switch index {
	case 0:
		return "Unknown"
	case 1:
		return "Charging"
	case 2:
		return "Discharging"
	case 3:
		return "Empty"
	case 4:
		return "Fully charged"
	case 5:
		return "Pending charge"
	case 6:
		return "Pending discharge"
	default:
		return "Unknown"
	}
}

func deviceTecnology(index uint32) string {
	switch index {
	case 0:
		return "Unknown"
	case 1:
		return "Lithium ion"
	case 2:
		return "Lithium polymer"
	case 3:
		return "Lithium iron phosphate"
	case 4:
		return "Lead acid"
	case 5:
		return "Nickel cadmium"
	case 6:
		return "Nickel metal hydride"
	default:
		return "Unknown"
	}
}
