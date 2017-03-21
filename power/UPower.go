package main

import (
	"github.com/godbus/dbus"
	"net/http"
	"github.com/surlykke/RefudeServices/common"
	"github.com/surlykke/RefudeServices/service"
)

type UPowerObject interface {
	service.Resource
	ReadDBusProps(m map[string]dbus.Variant)
	Copy() UPowerObject
}


type UPower struct {
    DaemonVersion  string
    CanSuspend     bool
    CanHibernate   bool
    OnBattery      bool
    OnLowBattery   bool
    LidIsClosed    bool
    LidIsPresent   bool
}

type PropertyObject interface {
	ReadDBusProps(m map[string]dbus.Variant)
}


func (up *UPower) ReadDBusProps(m map[string]dbus.Variant) {
	for key,variant := range m {
		switch key {
		case "DaemonVersion": up.DaemonVersion = variant.Value().(string)
		case "CanSuspend": up.CanSuspend = variant.Value().(bool)
		case "CanHibernate": up.CanHibernate = variant.Value().(bool)
		case "OnBattery": up.OnBattery = variant.Value().(bool)
		case "OnLowBattery": up.OnLowBattery = variant.Value().(bool)
		case "LidIsClosed": up.LidIsClosed = variant.Value().(bool)
		case "LidIsPresent": up.LidIsPresent = variant.Value().(bool)
		}
	}
}

func (up UPower) Data(r *http.Request) (int, string, []byte) {
	return common.GetJsonData(up)
}

type Device struct {
	NativePath       string
	Vendor           string
	Model            string
	Serial           string
	UpdateTime       uint64
	Type             uint32
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
	State            uint32
	IsRechargeable   bool
	Capacity         float64
	Technology       uint32
}


func (d *Device) ReadDBusProps(m map[string]dbus.Variant) {
	for key, variant := range m {
		switch key {
		case "NativePath": d.NativePath = variant.Value().(string)
		case "Vendor": d.Vendor = variant.Value().(string)
		case "Model": d.Model = variant.Value().(string)
		case "Serial": d.Serial = variant.Value().(string)
		case "UpdateTime": d.UpdateTime = variant.Value().(uint64)
		case "Type": d.Type = variant.Value().(uint32)
		case "PowerSupply": d.PowerSupply = variant.Value().(bool)
		case "HasHistory": d.HasHistory = variant.Value().(bool)
		case "HasStatistics": d.HasStatistics = variant.Value().(bool)
		case "Online": d.Online = variant.Value().(bool)
		case "Energy": d.Energy = variant.Value().(float64)
		case "EnergyEmpty": d.EnergyEmpty = variant.Value().(float64)
		case "EnergyFull": d.EnergyFull = variant.Value().(float64)
		case "EnergyFullDesign": d.EnergyFullDesign = variant.Value().(float64)
		case "EnergyRate": d.EnergyRate = variant.Value().(float64)
		case "Voltage": d.Voltage = variant.Value().(float64)
		case "TimeToEmpty": d.TimeToEmpty = variant.Value().(int64)
		case "TimeToFull": d.TimeToFull = variant.Value().(int64)
		case "Percentage": d.Percentage = variant.Value().(float64)
		case "IsPresent": d.IsPresent = variant.Value().(bool)
		case "State": d.State = variant.Value().(uint32)
		case "IsRechargeable": d.IsRechargeable = variant.Value().(bool)
		case "Capacity": d.Capacity = variant.Value().(float64)
		case "Technology": d.Technology = variant.Value().(uint32)
		}
	}
}

func (d Device) Data(r *http.Request) (int, string, []byte) {
	return common.GetJsonData(d)
}

