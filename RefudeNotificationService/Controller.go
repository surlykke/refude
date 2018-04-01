// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/godbus/dbus"
	"errors"
	"fmt"
	"github.com/godbus/dbus/introspect"
	"github.com/surlykke/RefudeServices/lib/service"
	"time"
	"strings"
	"strconv"
	"net/url"
)

const NOTIFICATIONS_SERVICE = "org.freedesktop.Notifications"
const NOTIFICATIONS_PATH = "/org/freedesktop/Notifications"
const NOTIFICATIONS_INTERFACE = NOTIFICATIONS_SERVICE
const INTROSPECT_INTERFACE = "org.freedesktop.DBus.Introspectable"
const INTROSPECT_XML =
`<!DOCTYPE node PUBLIC "-//freedesktop//DTD D-BUS Object Introspection 1.0//EN"
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


var	conn *dbus.Conn
var ids = make(chan uint32, 0)

const (
	Expired uint32 = 1
	Dismissed = 2
	Closed = 3
)

func generate(out chan uint32) {
	for id := uint32(1); ; id++ {
		out <- id
	}
}

func closeNotificationAfter(path string, eTag string, milliseconds int32) {
	time.Sleep(time.Duration(milliseconds)*time.Millisecond)
	service.UnMapIfMatch(path, eTag)
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
	fmt.Println("Notify:", app_name, replaces_id, app_icon, summary, body, actions, hints, expire_timeout)
	id := replaces_id
	if id == 0 {
		id = <- ids
	}

	path := fmt.Sprintf("/notifications/%d", id)

	notification := Notification{
		Id : id,
		Sender: app_name,
		Subject: sanitize(summary, []string{}, []string{}),
		Body: sanitize(body, allowedTags, allowedEscapes),
		Actions: map[string]string{},
		eTag : fmt.Sprintf("%d", id),
		Self : "notifications-service:" + path,
	}

	for i := 0; i + 1 < len(actions); i = i + 2 {
		notification.Actions[actions[i]] = actions[i + 1]
	}


	service.Map( path, &notification)

	if expire_timeout == 0 {
		expire_timeout = 2000
	}

	if expire_timeout > 0 {
		go closeNotificationAfter(path, notification.eTag, expire_timeout)
	}

	return id, nil
}

func CloseNotification(id uint32) *dbus.Error {
	path := fmt.Sprintf("/notifications/%d", id)
	service.Unmap(path)
	return nil
}

func close(path string, eTag string, reason uint32) {
	if reason == Expired {
		if service.UnMapIfMatch(path, eTag) {
			notificationClosed(idFromPath(path), reason)
		}
	} else {
		if service.Unmap(path) {
			notificationClosed(idFromPath(path), reason)
		}
	}
}

func idFromPath(path string) uint32 {
	idS := path[len("/notifications/"):]
	if id, err := strconv.ParseUint(idS, 10, 32); err != nil {
		panic("Invalid id: " + idS)
	} else  {
		return uint32(id)
	}
}

func notificationClosed(id uint32, reason uint32) {
	conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE + ".NotificationClosed", id, reason)
}

func GetServerInformation() (string, string, string, string, *dbus.Error) {
	return "Refude", "Refude", "0.1-alpha", "1.2", nil
}


var allowedEscapes =
	[]string{ "&amp;", "&#38;", "&#x26;", "&lt;", "&#60;", "&#x3C;", "&#x3c;", "&gt;", "&#62;", "&#x3E;", "&#x3e;", "&apos;", "&quot;"}

var allowedTags =
	[]string{"<b>", "</b>", "<i>", "</i>", "<u>", "</u>"}

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
	*src = (*src)[endMarkerPos + 1:]
}


func Setup() {
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
	conn.ExportMethodTable(
		map[string]interface{}{
			"GetCapabilities": GetCapabilities,
			"Notify": Notify,
			"CloseNotification": CloseNotification,
			"GetServerInformation": GetServerInformation,
		},
		NOTIFICATIONS_PATH,
		NOTIFICATIONS_INTERFACE,
	)
	conn.Export(introspect.Introspectable(INTROSPECT_XML), NOTIFICATIONS_PATH, INTROSPECT_INTERFACE)
}

func filterMethod(resource interface{}, query url.Values) bool {
	if n, ok := resource.(*Notification); ok {
		if searchTerms, ok := query["q"]; ok {
			for _,searchTerm := range searchTerms {
				if strings.Contains(strings.ToUpper(n.Subject), strings.ToUpper(searchTerm)) ||
				   strings.Contains(strings.ToUpper(n.Body), strings.ToUpper(searchTerm)) {
					return true
				}
			}
		}
	}
	return false
}
