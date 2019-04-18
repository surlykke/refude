package power

import (
	"github.com/godbus/dbus"
	dbuscall "github.com/surlykke/RefudeServices/lib/dbusutils"
	"github.com/surlykke/RefudeServices/lib/resource"
)

const UPowService = "org.freedesktop.UPower"
const UPowPath = "/org/freedesktop/UPower"
const UPowerInterface = "org.freedesktop.UPower"
const DevicePrefix = "/org/freedesktop/UPower/devices"
const DisplayDevicePath = DevicePrefix + "/DisplayDevice"
const UPowerDeviceInterface = "org.freedesktop.UPower.Device"
const login1Service = "org.freedesktop.login1"
const login1Path = "/org/freedesktop/login1"
const managerInterface = "org.freedesktop.login1.Manager"

func setup() chan *dbus.Signal {
	var signals = make(chan *dbus.Signal, 100)

	dbusConn.Signal(signals)
	dbusConn.BusObject().Call(
		"org.freedesktop.DBus.AddMatch",
		0,
		"type='signal',interface='org.freedesktop.DBus.Properties',member='PropertiesChanged', sender='org.freedesktop.UPower'")

	enumCall := dbusConn.Object(UPowService, UPowPath).Call(UPowerInterface+".EnumerateDevices", dbus.Flags(0))
	devicePaths := append(enumCall.Body[0].([]dbus.ObjectPath), DisplayDevicePath)
	for _, path := range devicePaths {
		var device = &Device{}
		device.DisplayDevice = path == DisplayDevicePath
		device.GenericResource = resource.MakeGenericResource(deviceSelf(path), DeviceMediaType)
		device.DbusPath = path
		updateDevice(device, dbuscall.GetAllProps(dbusConn, UPowService, path, UPowerDeviceInterface))
		setDevice(device)
	}

	return signals
}

var dbusConn = func() *dbus.Conn {
	if conn, err := dbus.SystemBus(); err != nil {
		panic(err)
	} else {
		return conn
	}
}()

var possibleActionValues = map[string][]string{
	"PowerOff":    {"Shutdown", "Power off the machine", "system-shutdown"},
	"Reboot":      {"Reboot", "Reboot the machine", "system-reboot"},
	"Suspend":     {"Suspend", "Suspend the machine", "system-suspend"},
	"Hibernate":   {"Hibernate", "Put the machine into hibernation", "system-suspend-hibernate"},
	"HybridSleep": {"HybridSleep", "Put the machine into hybrid sleep", "system-suspend-hibernate"}}

func buildSessionResource() *resource.GenericResource {
	var session = resource.MakeGenericResource("/session", SessionMediaType)
	for id, pv := range possibleActionValues {
		if "yes" == dbusConn.Object(login1Service, login1Path).Call(managerInterface+".Can"+id, dbus.Flags(0)).Body[0].(string) {
			var dbusEndPoint = managerInterface + "." + id
			var executer = func() {
				dbusConn.Object(login1Service, login1Path).Call(dbusEndPoint, dbus.Flags(0), false)
			}
			session.ResourceActions[id] = resource.ResourceAction{Description: pv[1], IconName: pv[2], Executer: executer}
		}
	}

	return &session
}

func updateDevice(d *Device, m map[string]dbus.Variant) {
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
