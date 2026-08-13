// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package notifications

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/introspect"
	"github.com/surlykke/refude/internal/lib/image"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/xdg"
	"github.com/surlykke/refude/pkg/pubsub"
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
	Expired uint32 = iota + 1
	Dismissed
	Closed
)

var conn *dbus.Conn
var ids = make(chan uint32)

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

func Notify(
	app_name string,
	replaces_id uint32,
	app_icon string,
	summary string,
	body string,
	actions []string,
	hints map[string]dbus.Variant,
	expire_timeout int32) (uint32,
	*dbus.Error) {

	var id uint32
	if replaces_id != 0 {
		id = replaces_id
	} else {
		id = <-ids
	}

	// Get image

	var iconName string
	if variant, ok := hints["image-data"]; ok {
		if name, ok := installRawImageIcon(variant); ok {
			iconName = name
		}
	} else if variant, ok := hints["image_data"]; ok {
		if name, ok := installRawImageIcon(variant); ok {
			iconName = name
		}
	} else if variant, ok := hints["image-path"]; ok {
		if path, ok := variant.Value().(string); ok {
			iconName = path
		}
	} else if variant, ok := hints["image_path"]; ok {
		if path, ok := variant.Value().(string); ok {
			iconName = path
		}
	} else if app_icon != "" {
		iconName = app_icon
	} else if variant, ok := hints["icon_data"]; ok {
		if name, ok := installRawImageIcon(variant); ok {
			iconName = name
		}
	}

	var title = sanitize(summary, []string{}, []string{})
	body = sanitize(body, allowedTags, allowedEscapes)
	notification := Notification{
		Base:           resource.MakeBase(id2Path(id), title, app_name+": "+body, iconName, "Notification"),
		NotificationId: id,
		Sender:         app_name,
		Created:        time.Now(),
		Body:           body,
		Urgency:        Normal,
		NActions:       map[string]string{},
		Hints:          map[string]any{},
		IconName:       iconName,
	}

	for i := 0; i+1 < len(actions); i = i + 2 {
		notification.NActions[actions[i]] = actions[i+1]
		notification.AddAction(actions[i+1], "", resource.OrgRefudeAction, "action="+actions[i])
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
				log.Print("urgency hint not of type uint8, rather:", reflect.TypeOf(val.Value()))
			}
		}
		if acceptableHintTypes[val.Signature().String()] {
			notification.Hints[name] = val.Value()
		}
	}

	switch notification.Urgency {
	case Low:
		if expire_timeout >= 2_000 || expire_timeout < 0 {
			expire_timeout = 2_000
		}
	case Normal:
		if expire_timeout >= 10_000 || expire_timeout < 0 {
			expire_timeout = 10_000
		}
	default:
		expire_timeout = 3_600_000
	}

	notification.Expires = time.Now().Add(time.Duration(expire_timeout) * time.Millisecond)
	newNotifications <- &notification

	return id, nil
}

func installRawImageIcon(v dbus.Variant) (string, bool) {
	if path, err := extractRawImageAndSaveAsPng(v); err != nil {
		log.Println("Unable to install raw image", err)
		return "", false
	} else {
		return path, true
	}

}

func extractRawImageAndSaveAsPng(v dbus.Variant) (string, error) {
	var id image.ImageData

	// I'll never be a fan of dbus...
	if v.Signature().String() != "av" {
		return "", errors.New("Not an array of variants")
	} else if ifarray, ok := v.Value().([]any); !ok {
		return "", errors.New("Value not an array of interface{}")
	} else if len(ifarray) != 7 {
		return "", errors.New("len not 7")
	} else if id.Width, ok = ifarray[0].(int32); !ok {
		return "", errors.New("arr[0] not an int32")
	} else if id.Height, ok = ifarray[1].(int32); !ok {
		return "", errors.New("arr[1] not an int32")
	} else if id.Rowstride, ok = ifarray[2].(int32); !ok {
		return "", errors.New("arr[2] not an int32")
	} else if id.HasAlpha, ok = ifarray[3].(bool); !ok {
		return "", errors.New("arr[3] not a bool")
	} else if id.BitsPrSample, ok = ifarray[4].(int32); !ok {
		return "", errors.New("arr[4] not an int32")
	} else if id.Channels, ok = ifarray[5].(int32); !ok {
		return "", errors.New("arr[5] not an int32")
	} else if id.Data, ok = ifarray[6].([]uint8); !ok {
		return "", errors.New("arr[6] not an []uint8")
	} else {
		var filePath = xdg.SessionFilesDir + "/" + image.ImageDataHashName(id) + ".png"
		if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
			if bytes, err := id.AsPng(); err != nil {
				return "", err
			} else if err := os.WriteFile(filePath, bytes, 0700); err != nil {
				os.Remove(filePath)
				return "", err
			}
		}
		return filePath, nil
	}
}

func CloseNotification(id uint32) {
	closedNotifications <- id
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

var NotificationEvents = pubsub.MakePublisher[struct{}]()
var newNotifications = make(chan *Notification)
var closedNotifications = make(chan uint32)
var dismissedNotifications = make(chan uint32)

func Run(noNotifications bool) {
	if noNotifications {
		return
	}

	var err error
	var reply dbus.RequestNameReply

	defer func() {
		if err := recover(); err != nil {
			log.Print(err, "- hence Notifications not running")
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
		map[string]any{
			"GetCapabilities":      GetCapabilities,
			"Notify":               Notify,
			"CloseNotification":    CloseNotification,
			"GetServerInformation": GetServerInformation,
		},
		NOTIFICATIONS_PATH,
		NOTIFICATIONS_INTERFACE,
	)
	_ = conn.Export(introspect.Introspectable(INTROSPECT_XML), NOTIFICATIONS_PATH, INTROSPECT_INTERFACE)

	var ticker = time.NewTicker(500 * time.Millisecond)
	for {
		var somethingChanged = false
		select {
		case n := <-newNotifications:
			resource.Put(n)
			somethingChanged = true
		case id := <-closedNotifications:
			if _, ok := resource.Remove(id2Path(id)); ok {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, Closed)
				somethingChanged = true
			}
		case id := <-dismissedNotifications:
			if _, ok := resource.Remove(id2Path(id)); ok {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, Dismissed)
				somethingChanged = true
			}
		case <-ticker.C:
			for _, n := range resource.GetByPrefix[*Notification]("/notification/") {
				if !n.Expired && n.Expires.Before(time.Now()) {
					var copy = *n
					copy.Expired = true
					resource.Put(&copy)
					conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", copy.NotificationId, Expired)

					somethingChanged = true

				}
			}

		}
		if somethingChanged {
			NotificationEvents.Publish(struct{}{})
		}
	}
}

func id2Path(id uint32) string {
	return fmt.Sprintf("/notification/%d", id)
}
