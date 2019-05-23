// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package notifications

import (
	"bytes"
	"crypto/sha1"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/lib/xdg"

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

var imageDir = func() string {
	var dir = xdg.RuntimeDir + "/org.refude.notification-images/"
	if err := os.MkdirAll(dir, 0700); err != nil {
		panic(err)
	}
	return dir
}()

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

type ImageData struct {
	width        int32
	height       int32
	rowstride    int32
	hasAlpha     bool
	bitsPrSample int32
	channels     int32
	data         []uint8
}

func (id ImageData) AsPng() ([]byte, error) {
	if id.channels != 3 && id.channels != 4 {
		return nil, fmt.Errorf("Don't know how to deal with %d", id.channels)
	} else if id.channels == 4 && !id.hasAlpha {
		return nil, fmt.Errorf("hasAlpha, but not 4 channels")
	}

	pngData := image.NewRGBA(image.Rect(0, 0, int(id.width), int(id.height)))
	pixelStride := id.rowstride / id.width
	fmt.Println("rowstride:", id.rowstride, ", pixelstride:", pixelStride)
	var count = 0

	for y := int32(0); y < id.height; y++ {
		for x := int32(0); x < id.width; x++ {
			count++
			pos := int(y*id.rowstride + x*pixelStride)
			var alpha = uint8(255)
			if id.hasAlpha {
				alpha = id.data[pos+3]
			}
			pngData.Set(int(x), int(y), color.RGBA{R: id.data[pos], G: id.data[pos+1], B: id.data[pos+2], A: alpha})
		}
	}

	buf := &bytes.Buffer{}
	err := png.Encode(buf, pngData)
	if err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

func getRawImage(v dbus.Variant) string {
	var id ImageData
	var err error = nil

	// I'll never be a fan of dbus...
	if v.Signature().String() != "av" {
		err = fmt.Errorf("Not an array of variants")
	} else if ifarray, ok := v.Value().([]interface{}); !ok {
		err = fmt.Errorf("Value not an array of interface{}")
	} else if len(ifarray) != 7 {
		err = fmt.Errorf("len not 7")
	} else if id.width, ok = ifarray[0].(int32); !ok {
		err = fmt.Errorf("arr[0] not an int32")
	} else if id.height, ok = ifarray[1].(int32); !ok {
		err = fmt.Errorf("arr[1] not an int32")
	} else if id.rowstride, ok = ifarray[2].(int32); !ok {
		err = fmt.Errorf("arr[2] not an int32")
	} else if id.hasAlpha, ok = ifarray[3].(bool); !ok {
		err = fmt.Errorf("arr[3] not a bool")
	} else if id.bitsPrSample, ok = ifarray[4].(int32); !ok {
		err = fmt.Errorf("arr[4] not an int32")
	} else if id.channels, ok = ifarray[5].(int32); !ok {
		err = fmt.Errorf("arr[5] not an int32")
	} else if id.data, ok = ifarray[6].([]uint8); !ok {
		err = fmt.Errorf("arr[6] not an []uint8")
	}

	bytes, err := id.AsPng()

	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		var pngName = fmt.Sprintf("%X", sha1.Sum(bytes))
		var pngPath = fmt.Sprintf("%s%s.png", imageDir, pngName)
		if err = ioutil.WriteFile(pngPath, bytes, 0700); err != nil {
			fmt.Println(err)
			return ""
		} else {
			return imageDir + pngName + ".png"
		}
	}
}

func getFile(v dbus.Variant) string {
	var err error
	var pngName string

	if str, ok := v.Value().(string); !ok {
		err = fmt.Errorf("Value not a string")
	} else if bytes, err := ioutil.ReadFile(str); err == nil {
		pngName = fmt.Sprintf("%X", sha1.Sum(bytes))
		var pngPath = fmt.Sprintf("%s%s.png", imageDir, pngName)
		err = ioutil.WriteFile(pngPath, bytes, 0700)
	}

	if err != nil {
		fmt.Println(err)
		return ""
	} else {
		return imageDir + pngName + ".png"
	}
}

func notify(app_name string,
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

	notification := &Notification{}
	notification.Init(notificationSelf(id), "notification")
	notification.Id = id
	notification.Sender = app_name
	notification.Created = time.Now()
	notification.Subject = sanitize(summary, []string{}, []string{})
	notification.Body = sanitize(body, allowedTags, allowedEscapes)

	// Get image
	if v, ok := hints["image-data"]; ok {
		notification.imagePath = getRawImage(v)
	} else if v, ok := hints["image_data"]; ok {
		notification.imagePath = getRawImage(v)
	} else if v, ok := hints["image-path"]; ok {
		notification.imagePath = getFile(v)
	} else if v, ok := hints["image_path"]; ok {
		notification.imagePath = getFile(v)
	} else if v, ok := hints["icon_data"]; ok {
		notification.imagePath = getRawImage(v)
	}

	if expire_timeout == 0 {
		expire_timeout = 2000
	}

	if expire_timeout > 0 {
		notification.Expires = notification.Created.Add(time.Millisecond * time.Duration(expire_timeout))
	}

	time.AfterFunc(time.Minute*61, func() {
		reaper <- notification.Id
	})

	// Add a dismiss action
	var notificationId = notification.Id
	notification.SetDeleteAction(
		&resource.DeleteAction{
			Description: "dismiss",
			Executer:    func() { removals <- removal{notificationId, Dismissed} },
		})

	// Add actions given in notification (We are aware that one of these may overwrite the dismiss action added above)
	for i := 0; i+1 < len(actions); i = i + 2 {
		var notificationId = notification.Id
		var actionId = actions[i]
		var actionDescription = actions[i+1]
		notification.AddAction(actionId, resource.ResourceAction{
			Description: actionDescription, IconName: "", Executer: func() {
				conn.Emit(NOTIFICATIONS_PATH, NOTIFICATIONS_INTERFACE+".ActionInvoked", notificationId, actionId)
			},
		})
	}

	incomingNotifications <- notification
	return id, nil
}

func makeCloseFuntion(removals chan removal) interface{} {
	return func(id uint32) *dbus.Error {
		removals <- removal{id, Closed}
		return nil
	}
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
			"Notify":               notify,
			"CloseNotification":    makeCloseFuntion(removals),
			"GetServerInformation": GetServerInformation,
		},
		NOTIFICATIONS_PATH,
		NOTIFICATIONS_INTERFACE,
	)
	_ = conn.Export(introspect.Introspectable(INTROSPECT_XML), NOTIFICATIONS_PATH, INTROSPECT_INTERFACE)
}
