package power

import (
	"fmt"
	"os"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/godbus/dbus/v5/prop"
)

const WATCHER_SERVICE string = "org.kde.StatusNotifierWatcher"
const WATCHER_PATH dbus.ObjectPath = "/StatusNotifierWatcher"
const WATCHER_INTERFACE string = "org.kde.StatusNotifierWatcher"

const DBUS_SNI_SERVICE_NAME string = "org.refude.battery"
const DBUS_SNI_OBJECT_PATH dbus.ObjectPath = "/org/refude/battery"
const SNI_INTERFACE = "org.kde.StatusNotifierItem"

type BatteryObject struct{}

func (this BatteryObject) Scroll(delta int32) *dbus.Error {
	return nil
}

func (this BatteryObject) SecondaryActivate(x, y int32) *dbus.Error {
	return nil
}

func (this BatteryObject) XAyatanaSecondaryActivate(timestamp uint32) *dbus.Error {
	return nil
}

func emitter(conn *dbus.Conn, signal string) func(change *prop.Change) *dbus.Error {
	return func(change *prop.Change) *dbus.Error {
		conn.Emit(DBUS_SNI_OBJECT_PATH, SNI_INTERFACE+"."+signal)
		return nil
	}
}

func propOf(val string, writeable bool, emit prop.EmitType, callBack func(change *prop.Change) *dbus.Error) *prop.Prop {
	return &prop.Prop{
		Value:    val,
		Writable: writeable,
		Emit:     emit,
		Callback: callBack,
	}
}

func tray_applet_run() {
	fmt.Println("tray_applet_run")

	var (
		err error
	)

	var conn *dbus.Conn
	if conn, err = dbus.ConnectSessionBus(); err != nil {
		panic(err)
	}
	defer conn.Close()

	var reply dbus.RequestNameReply
	if reply, err = conn.RequestName(DBUS_SNI_SERVICE_NAME, dbus.NameFlagDoNotQueue); err != nil {
		panic(err)
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		panic("name already taken")
	}

	var batteryObject BatteryObject
	if err = conn.Export(batteryObject, DBUS_SNI_OBJECT_PATH, SNI_INTERFACE); err != nil {
		panic(err)
	}

	var propsSpec = map[string]map[string]*prop.Prop{
		SNI_INTERFACE: {
			"Id":                          propOf(fmt.Sprintf("%s-%d", DBUS_SNI_SERVICE_NAME, os.Getpid()), false, prop.EmitTrue, nil),
			"Category":                    propOf("Hardware", false, prop.EmitTrue, nil),
			"Status":                      propOf("Active", false, prop.EmitTrue, nil),
			"IconName":                    propOf("", true, prop.EmitTrue, emitter(conn, "NewIcon")),
			"IconAccessibleDesc":          propOf("", false, prop.EmitTrue, nil),
			"AttentionIconName":           propOf("", false, prop.EmitTrue, nil),
			"AttentionIconAccessibleDesc": propOf("", false, prop.EmitTrue, nil),
			"Title":                       propOf("Battery", false, prop.EmitTrue, nil),
			"IconThemePath":               propOf("", false, prop.EmitTrue, nil),
			"Menu":                        propOf("", false, prop.EmitTrue, nil),
			"XAyatanaLabel":               propOf("", false, prop.EmitTrue, nil),
			"XAyatanaLabelGuide":          propOf("", false, prop.EmitTrue, nil),
			"XAyatanaOrderingIndex":       propOf("", false, prop.EmitTrue, nil),
		},
	}

	var props *prop.Properties
	if props, err = prop.Export(conn, DBUS_SNI_OBJECT_PATH, propsSpec); err != nil {
		panic(err)
	}

	var node *introspect.Node
	node = &introspect.Node{
		Name: string(DBUS_SNI_OBJECT_PATH),
		Interfaces: []introspect.Interface{
			introspect.IntrospectData,
			prop.IntrospectData,
			{
				Name:       SNI_INTERFACE,
				Methods:    introspect.Methods(batteryObject),
				Properties: props.Introspection(SNI_INTERFACE),
				Signals: []introspect.Signal{
					{Name: "NewIcon", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
					{Name: "NewAttentionIcon", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
					{Name: "NewIconThemePath", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
					{Name: "NewStatus", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
					{Name: "NewTitle", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
					{Name: "XayatanaNewLabel", Args: []introspect.Arg{}, Annotations: []introspect.Annotation{}},
				},
			},
		},
	}
	if err = conn.Export(introspect.NewIntrospectable(node), DBUS_SNI_OBJECT_PATH, "org.freedesktop.DBus.Introspectable"); err != nil {
		panic(err)
	}
	var call *dbus.Call
	if call = conn.Object(WATCHER_SERVICE, WATCHER_PATH).Call(WATCHER_INTERFACE+".RegisterStatusNotifierItem", dbus.Flags(0), string(DBUS_SNI_OBJECT_PATH)); call.Err != nil {
		panic(call.Err)
	}

	go watchBattery(props)

	c := make(chan *dbus.Signal)
	conn.Signal(c)
	for range c {
	}

}

func getIconName() string {
	var displayDevice, _ = DeviceMap.Get("DisplayDevice")
	var state, percentage = displayDevice.State, displayDevice.Percentage
	if state == "Discharging" || state == "Empty" {
		state = "discharging"
	} else {
		state = "charging"
	}
	return fmt.Sprintf("refude_battery_%s_%d0", state, int(percentage+5)/10)
}

func watchBattery(props *prop.Properties) {
	var prevIcon = ""
	var subscription = DeviceMap.Events.Subscribe()
	for {
		var iconName = getIconName()
		if iconName != prevIcon {
			prevIcon = iconName
			if err := props.Set(SNI_INTERFACE, "IconName", dbus.MakeVariant(iconName)); err != nil {
				fmt.Println("Error setting IconName:", err)
			}
		}
		subscription.Next()
	}
}
