// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/icons"
	"github.com/godbus/dbus"
	"log"
	"strconv"
	"github.com/surlykke/RefudeServices/lib/resource"
	"fmt"
	"reflect"
)


type Item struct {
	Id string
	Category string
	Status string
	IconName string
	IconAccessibleDesc string
	AttentionIconName string
	AttentionAccessibleDesc string
	Title string
	Menu *MenuItem `json:"omitempty"`

	iconThemePath string

	menu dbus.BusObject
	dbusObj dbus.BusObject
	path string
}

type MenuItem struct {

}


func (item *Item) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(item, w)
}

func (item *Item) POST(w http.ResponseWriter, r *http.Request) {
	method := resource.GetSingleQueryParameter(r, "method", "Activate")
	x, errX := strconv.Atoi(resource.GetSingleQueryParameter(r, "x", "0"))
	y, errY := strconv.Atoi(resource.GetSingleQueryParameter(r, "y", "0"))
	if (method != "Activate" && method != "SecondaryActivate" && method != "ContextMenu") ||
		errX != nil ||
		errY != nil {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	item.dbusObj.Call("org.kde.StatusNotifierItem."+method, dbus.Flags(0), x, y)
	w.WriteHeader(http.StatusAccepted)
}

func (item *Item) copy() *Item {
	var tmp = *item
	return &tmp
}

func (item *Item) getProp(propName string) string {
	var method = PROPERTIES_INTERFACE + ".Get"
	if call := item.dbusObj.Call(method, dbus.Flags(0), ITEM_INTERFACE, propName);  call.Err != nil {
		log.Println(call.Err)
		return ""
	} else if value, ok := call.Body[0].(dbus.Variant).Value().(string); !ok {
		log.Printf("Property '%s' not of type string", propName)
		return ""
	} else {
		return value
	}
}

type SenderAndPath struct {
	sender  string
	objPath dbus.ObjectPath
}

func MakeItem(sp SenderAndPath) *Item {
	var item = &Item{}
	item.dbusObj = conn.Object(sp.sender, sp.objPath)
	item.Id = item.getProp("Id")
	item.Category = item.getProp("Category")
	item.Status = item.getProp("Status")
	item.iconThemePath = item.getProp("IconThemePath");
	item.IconName = item.getProp("IconName")
	if item.IconName == "" {
		item.IconName = item.getPixMap("IconPixmap")
	} else if item.iconThemePath != ""{
		icons.CopyIcons(item.IconName, item.iconThemePath)
	}
	item.IconAccessibleDesc = item.getProp("IconAccessibleDesc")
	item.AttentionIconName = item.getProp("AttentionIconName")
	if item.AttentionIconName == ""	{
		item.AttentionIconName = item.getPixMap("AttentionIconPixmap")
	} else if item.iconThemePath != "" {
		icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
	}
	item.AttentionAccessibleDesc = item.getProp("AttentionAccessibleDesc")
	item.Title = item.getProp("Title")


	return item
}

func (item *Item) Run(path string, signals chan *dbus.Signal) {
	// TODO Menu
	service.Map(path, item.copy())
	defer service.Unmap(path)

	for signal := range signals {
		fmt.Print(path, " signal: ", signal.Name, signal.Body, "\n")
		switch signal.Name {
		case "org.kde.StatusNotifierItem.NewIcon":
			item.IconName = item.getProp("IconName")
			item.IconAccessibleDesc = item.getProp("IconAccessibleDesc")
			if item.IconName == "" {
				item.IconName = item.getPixMap("IconPixmap")
			} else if item.iconThemePath != ""{
				icons.CopyIcons(item.IconName, item.iconThemePath)
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewAttentionIcon":
			item.AttentionIconName = item.getProp("AttentionIconName")
			item.AttentionAccessibleDesc = item.getProp("AttentionAccessibleDesc")
			if item.AttentionIconName == ""	{
				item.AttentionIconName = item.getPixMap("AttentionIconName")
			} else if item.iconThemePath != "" {
				icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewStatus":
			if tmp, ok := signal.Body[0].(string); ok {
				item.Status = tmp;
			} else {
				log.Println("NewStatus signal: ", signal.Body[0], ", not a string")
			}
			service.Map(path, item.copy())
		case "org.kde.StatusNotifierItem.NewIconThemePath":
			item.iconThemePath,_ = signal.Body[0].(string)
			if item.iconThemePath != "" {
				if item.IconName != "" {
					icons.CopyIcons(item.IconName, item.iconThemePath)
				}
				if item.AttentionIconName != "" {
					icons.CopyIcons(item.AttentionIconName, item.iconThemePath)
				}
			}
		}
	}
}

func (item *Item) getPixMap(propName string) string {
	if call := item.dbusObj.Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), ITEM_INTERFACE, propName); call.Err != nil {
		return ""
	} else {
		value := call.Body[0].(dbus.Variant).Value()
		dbusValue, ok := value.([][]interface{})
		if !ok {
			log.Println("Expected", propName, "to be of type [][]interface{}, but found", reflect.TypeOf(value))
			return ""
		}

		return collectPixMap(dbusValue)
	}

}

func collectPixMap(dbusValue [][]interface{}) string {
	res := make(icons.Icon, 0)
	for _, arr := range (dbusValue) {
		for len(arr) > 2 {
			width := arr[0].(int32)
			height := arr[1].(int32)
			pixels := arr[2].([]byte)
			res = append(res, icons.Img{Width: width, Height: height, Pixels: pixels})
			arr = arr[3:]
		}
	}

	return icons.SaveAsPngToSessionIconDir(res)
}
