// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"errors"
	"strings"
	"time"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
	"github.com/surlykke/RefudeServices/lib/resource"
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

const (
	Expired   uint32 = 1
	Dismissed        = 2
	Closed           = 3
)

type removal struct {
	id         uint32
	internalId uint32
	reason     uint32
}

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

func makeNotifyFunction(notifications chan *Notification) interface{} {
	return func(app_name string,
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

		notification := &Notification{
			Id:         id,
			internalId: <-ids,
			Sender:     app_name,
			Subject:    sanitize(summary, []string{}, []string{}),
			Body:       sanitize(body, allowedTags, allowedEscapes),
		}

		notification.AbstractResource = resource.MakeAbstractResource(notificationSelf(id), NotificationMediaType)

		if expire_timeout == 0 {
			expire_timeout = 2000
		}

		if expire_timeout > 0 {
			var timeToExpire = time.Millisecond * time.Duration(expire_timeout)
			var expires = time.Now().Add(timeToExpire)
			notification.Expires = &expires
			notification.removeAfter(timeToExpire)
		}

		// Add a dismiss action
		var notificationId = notification.Id
		notification.ResourceActions["dismiss"] = resource.ResourceAction{
			Description: "Dismiss", IconName: "", Executer: func() { removals <- removal{notificationId, 0, Dismissed} },
		}

		// Add actions given in notification (We are aware that one of these may overwrite the dismiss action added above)
		for i := 0; i+1 < len(actions); i = i + 2 {
			var notificationId = notification.Id
			var actionId = actions[i]
			var actionDescription = actions[i+1]
			notification.ResourceActions[actionId] = resource.ResourceAction{
				Description: actionDescription, IconName: "", Executer: func() {
					conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", notificationId, actionId)
				},
			}
		}

		notifications <- notification
		return id, nil
	}
}

func makeCloseFuntion(removals chan removal) interface{} {
	return func(id uint32) *dbus.Error {
		removals <- removal{id, 0, Closed}
		return nil
	}
}

func notificationClosed(id uint32, reason uint32) {
	conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".NotificationClosed", id, reason)
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

func DoDBus(notifications chan *Notification, removals chan removal) {
	var err error

	// Get on the bus
	conn, err = dbus.SessionBus()
	if err != nil {
		panic(err)
	} else if reply, err := conn.RequestName(NOTIFICATIONS_SERVICE, dbus.NameFlagDoNotQueue); err != nil {
		panic(err)
	} else if reply != dbus.RequestNameReplyPrimaryOwner {
		panic(errors.New(NOTIFICATIONS_SERVICE + " taken"))
	}

	go generate(ids)

	// Put StatusNotifierWatcher object up
	_ = conn.ExportMethodTable(
		map[string]interface{}{
			"GetCapabilities":      GetCapabilities,
			"Notify":               makeNotifyFunction(notifications),
			"CloseNotification":    makeCloseFuntion(removals),
			"GetServerInformation": GetServerInformation,
		},
		NOTIFICATIONS_PATH,
		NOTIFICATIONS_INTERFACE,
	)
	_ = conn.Export(introspect.Introspectable(INTROSPECT_XML), NOTIFICATIONS_PATH, INTROSPECT_INTERFACE)
}
