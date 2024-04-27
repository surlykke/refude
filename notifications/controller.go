// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package notifications

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
)

const NOTIFICATIONS_SERVICE = "org.freedesktop.Notifications"
const NOTIFICATIONS_PATH = "/org/freedesktop/Notifications"
const NOTIFICATIONS_INTERFACE = NOTIFICATIONS_SERVICE
const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const INTROSPECT_XML = `<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
        "http://www.freedesktop.org/standards/dbus/1.0/introspect.dtd">
<node>
    <interface name="org.freedesktop.DBus.Properties">
        <method name="Get">
            <arg type="s" name="interface_name" direction="in"/>
            <arg type="s" name="property_name" direction="in"/>
            <arg type="v" name="value" direction="out"/>
        </method>
        <method name="GetAll">
            <arg type="s" name="interface_name" direction="in"/>
            <arg type="a{sv}" name="properties" direction="out"/>
        </method>
        <method name="Set">
            <arg type="s" name="interface_name" direction="in"/>
            <arg type="s" name="property_name" direction="in"/>
            <arg type="v" name="value" direction="in"/>
        </method>
        <signal name="PropertiesChanged">
            <arg type="s" name="interface_name"/>
            <arg type="a{sv}" name="changed_properties"/>
            <arg type="as" name="invalidated_properties"/>
        </signal>
    </interface>
    <interface name="org.freedesktop.DBus.Introspectable">
        <method name="Introspect">
            <arg type="s" name="xml_data" direction="out"/>
        </method>
    </interface>
    <interface name="org.freedesktop.DBus.Peer">
        <method name="Ping"/>
        <method name="GetMachineId">
            <arg type="s" name="machine_uuid" direction="out"/>
        </method>
    </interface>
    <interface name="org.freedesktop.Notifications">
        <method name="GetCapabilities">
            <arg type="as" name="capabilities" direction="out"/>
        </method>
        <method name="Notify">
            <arg type="s" name="app_name" direction="in"/>
            <arg type="u" name="replaces_id" direction="in"/>
            <arg type="s" name="app_icon" direction="in"/>
            <arg type="s" name="summary" direction="in"/>
            <arg type="s" name="body" direction="in"/>
            <arg type="as" name="actions" direction="in"/>
            <arg type="a{sv}" name="hints" direction="in"/>
            <arg type="i" name="expire_timeout" direction="in"/>
            <arg type="u" name="id" direction="out"/>
        </method>
        <method name="CloseNotification">
            <arg type="u" name="id" direction="in"/>
        </method>
        <method name="GetServerInformation">
            <arg type="s" name="name" direction="out"/>
            <arg type="s" name="vendor" direction="out"/>
            <arg type="s" name="version" direction="out"/>
            <arg type="s" name="spec_version" direction="out"/>
        </method>
        <signal name="NotificationClosed">
            <arg type="u" name="id"/>
            <arg type="u" name="reason"/>
        </signal>
        <signal name="ActionInvoked">
            <arg type="u" name="id"/>
            <arg type="s" name="action_key"/>
        </signal>
    </interface>
</node>`

var acceptableHintTypes = map[string]bool{
	"y": true,
	"b": true,
	"n": true,
	"q": true,
	"i": true,
	"u": true,
	"x": true,
	"t": true,
	"d": true,
	"s": true,
	"o": true,
}

const (
	Expired   uint32 = 1
	Dismissed        = 2
	Closed           = 3
)

type removal struct{ id, reason uint32 }

var conn *dbus.Conn
var ids = make(chan uint32, 0)

func generate(out chan uint32) {
	for id := uint32(1); ; id++ {
		out <- id
	}
}

func GetCapabilities() ([]string, *dbus.Error) {
	return []string{
			"actions",
			"body",
			"body-hyperlinks",
			"body-markup",
			"icon-static",
		},
		nil
}

func Notify(app_name string,
	replaces_id uint32,
	app_icon string,
	summary string,
	body string,
	actions []string,
	hints map[string]dbus.Variant,
	expire_timeout int32) (uint32, *dbus.Error) {

	var id uint32
	if replaces_id != 0 {
		id = replaces_id
	} else {
		id = <-ids
	}

	// Get image

	var iconName string
	var ok bool

	var sizeHint uint32
	if iconName, sizeHint, ok = installRawImageIcon(hints, "image-data"); !ok {
		if iconName, sizeHint, ok = installRawImageIcon(hints, "image_data"); !ok {
			if iconName, ok = installFileIcon(hints, "image-path"); !ok {
				if iconName, ok = installFileIcon(hints, "image_path"); !ok {
					if "" != app_icon {
						iconName = app_icon
					} else {
						iconName, sizeHint, _ = installRawImageIcon(hints, "icon_data")
					}
				}
			}
		}
	}

	var title = sanitize(summary, []string{}, []string{})
	body = sanitize(body, allowedTags, allowedEscapes)

	notification := Notification{
		BaseResource: *resource.MakeBase(fmt.Sprintf("/notification/%d", id), title, body, iconName, "notification"),
		NotificationId: id,
		Sender:         app_name,
		Created:        UnixTime(time.Now()),
		Urgency:        Normal,
		NActions:       map[string]string{},
		Hints:          map[string]interface{}{},
		iconName:       iconName,
		IconSize:       sizeHint,
	}

	for i := 0; i+1 < len(actions); i = i + 2 {
		notification.NActions[actions[i]] = actions[i+1]
	}

	for name, val := range hints {
		if name == "urgency" {
			if b, ok := val.Value().(uint8); ok {
				if b == 0 {
					notification.Urgency = Low
				} else if b > 1 {
					notification.Urgency = Critical
				}
			} else {
				log.Info("urgency hint not of type uint8, rather:", reflect.TypeOf(val.Value()))
			}
		}
		if acceptableHintTypes[val.Signature().String()] {
			notification.Hints[name] = val.Value()
		}
	}

	notification.Expires = UnixTime(time.Now().Add(time.Duration(expire_timeout) * time.Millisecond))
	resourcerepo.Put(&notification)	
	calculateFlash()
	Updated.Store(time.Now().UnixMicro())	
	return id, nil
}

func installRawImageIcon(hints map[string]dbus.Variant, key string) (string, uint32, bool) {
	if v, ok := hints[key]; !ok {
		return "", 0, false
	} else if imageData, err := getRawImage(v); err != nil {
		log.Warn("Error converting variant to image data:", err)
		return "", 0, false
	} else {
		return icons.AddRawImageIcon(imageData), uint32(imageData.Width), true
	}
}

func getRawImage(v dbus.Variant) (image.ImageData, error) {
	var id image.ImageData
	var err error = nil

	// I'll never be a fan of dbus...
	if v.Signature().String() != "av" {
		return id, errors.New("Not an array of variants")
	} else if ifarray, ok := v.Value().([]interface{}); !ok {
		return id, errors.New("Value not an array of interface{}")
	} else if len(ifarray) != 7 {
		return id, errors.New("len not 7")
	} else if id.Width, ok = ifarray[0].(int32); !ok {
		return id, errors.New("arr[0] not an int32")
	} else if id.Height, ok = ifarray[1].(int32); !ok {
		return id, errors.New("arr[1] not an int32")
	} else if id.Rowstride, ok = ifarray[2].(int32); !ok {
		return id, errors.New("arr[2] not an int32")
	} else if id.HasAlpha, ok = ifarray[3].(bool); !ok {
		return id, errors.New("arr[3] not a bool")
	} else if id.BitsPrSample, ok = ifarray[4].(int32); !ok {
		return id, errors.New("arr[4] not an int32")
	} else if id.Channels, ok = ifarray[5].(int32); !ok {
		return id, errors.New("arr[5] not an int32")
	} else if id.Data, ok = ifarray[6].([]uint8); !ok {
		return id, errors.New("arr[6] not an []uint8")
	} else {
		return id, err
	}
}

func installFileIcon(hints map[string]dbus.Variant, key string) (string, bool) {
	if v, ok := hints[key]; !ok {
		return "", false
	} else if path, ok := v.Value().(string); !ok {
		log.Warn("Value not a string")
		return "", true
	} else {
		icons.AddFileIcon(path)
		return path, false
	}
}

func CloseNotification(id uint32) {
	removeNotification(id, Closed)
}

func GetServerInformation() (string, string, string, string, *dbus.Error) {
	return "Refude", "Refude", "0.1-alpha", "1.2", nil
}

var allowedEscapes = []string{"&amp;", "&#38;", "&#x26;", "&lt;", "&#60;", "&#x3C;", "&#x3c;", "&gt;", "&#62;", "&#x3E;", "&#x3e;", "&apos;", "&quot;"}

var allowedTags = []string{"<b>", "</b>", "<i>", "</i>", "<u>", "</u>"}

func sanitize(text string, allowedTags []string, allowedEscapes []string) string {
	sanitized := ""
	for len(text) > 0 {
		switch text[0:1] {
		case "<":
			helper(&text, &sanitized, allowedTags, ">")
		case "&":
			helper(&text, &sanitized, allowedEscapes, ";")
		default:
			sanitized += text[0:1]
			text = text[1:]
		}
	}
	return sanitized
}

func helper(src *string, dest *string, allowedPrefixes []string, endMarker string) {
	for _, prefix := range allowedPrefixes {
		if strings.HasPrefix(*src, prefix) {
			*dest += prefix
			*src = (*src)[len(prefix):]
			return
		}
	}
	endMarkerPos := strings.Index(*src, endMarker)
	if endMarkerPos < 0 {
		endMarkerPos = len(*src) - 1
	}
	*src = (*src)[endMarkerPos+1:]
}

func DoDBus() {
	var err error
	var reply dbus.RequestNameReply

	defer func() {
		if err := recover(); err != nil {
			log.Warn(err, "- hence Notifications not running")
		}
	}()

	// Get on the bus
	conn, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	}
	if reply, err = conn.RequestName(NOTIFICATIONS_SERVICE, dbus.NameFlagDoNotQueue); err != nil {
		panic(err)
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		panic(errors.New(NOTIFICATIONS_SERVICE + " taken"))
	}

	go generate(ids)

	// Put StatusNotifierWatcher object up
	_ = conn.ExportMethodTable(
		map[string]interface{}{
			"GetCapabilities":      GetCapabilities,
			"Notify":               Notify,
			"CloseNotification":    CloseNotification,
			"GetServerInformation": GetServerInformation,
		},
		NOTIFICATIONS_PATH,
		NOTIFICATIONS_INTERFACE,
	)
	_ = conn.Export(introspect.Introspectable(INTROSPECT_XML), NOTIFICATIONS_PATH, INTROSPECT_INTERFACE)
}
