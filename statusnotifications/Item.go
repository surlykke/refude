// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package statusnotifications

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/godbus/dbus/v5"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type Item struct {
	sender                  string
	path                    dbus.ObjectPath
	Id                      string
	Category                string
	Status                  string
	IconName                string
	IconAccessibleDesc      string
	AttentionIconName       string
	AttentionAccessibleDesc string
	Title                   string
	ToolTip                 string
	MenuPath                dbus.ObjectPath
	IconThemePath           string
	UseIconPixmap           bool
	UseAttentionIconPixmap  bool
	UseOverlayIconPixmap    bool
}

func (item *Item) Links(path string) link.List {
	if item.MenuPath != "" {
		return link.List{link.Make("/itemmenu/"+pathEscape(item.sender, item.path), "Menu", "", relation.Menu)}
	} else {
		return link.List{}
	}
}

func (item *Item) ForDisplay() bool {
	return false
}

func (item *Item) DoPost(w http.ResponseWriter, r *http.Request) {
	action := requests.GetSingleQueryParameter(r, "action", "left")
	x, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "x", "0"))
	y, _ := strconv.Atoi(requests.GetSingleQueryParameter(r, "y", "0"))

	var call *dbus.Call
	if slice.Among(action, "left", "middle", "right") {
		action2method := map[string]string{"left": "Activate", "middle": "SecondaryActivate", "right": "ContextMenu"}
		dbusObj := conn.Object(item.sender, item.path)
		call = dbusObj.Call("org.kde.StatusNotifierItem."+action2method[action], dbus.Flags(0), x, y)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}
	if call.Err != nil {
		log.Warn(call.Err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func collectPixMap(variant dbus.Variant) string {
	if arrs, ok := variant.Value().([][]interface{}); ok {
		var images = []image.ARGBImage{}
		for _, arr := range arrs {
			for len(arr) > 2 {
				width := uint32(arr[0].(int32))
				height := uint32(arr[1].(int32))
				pixels := arr[2].([]byte)
				images = append(images, image.ARGBImage{Width: width, Height: height, Pixels: pixels})
				arr = arr[3:]
			}
		}
		var argbIcon = image.ARGBIcon{Images: images}
		return icons.AddARGBIcon(argbIcon)
	}
	return ""
}

func pathEscape(sender string, path dbus.ObjectPath) string {
	return strings.ReplaceAll(sender+string(path), "/", "-")
}
