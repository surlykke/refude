// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/argb"
	"github.com/godbus/dbus"
	"reflect"
	"strings"
	"log"
	"strconv"
	"github.com/surlykke/RefudeServices/lib/resource"
	"path/filepath"
	"os"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io"
)

var propNames = []string{
	"Id",
	"Category",
	"Status",
	"Title",
	"ItemIsMenu",
	"IconName",
	"AttentionIconName",
	"OverlayIconName",
	"IconThemePath",
	"AttentionMovieName",
	"IconPixmap",
	"AttentionIconPixmap",
	"OverlayIconPixmap",
	//			"ToolTip",
}

type Item struct {
	props   map[string]interface{}
	dbusObj dbus.BusObject
}

func (item *Item) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(item.props, w)
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

func (item Item) fetchProps(propNames ...string) {
	for _, propName := range propNames {
		delete(item.props, propName)
		fmt.Println("Fetching", propName)
		call := item.dbusObj.Call(PROPERTIES_INTERFACE+".Get", dbus.Flags(0), ITEM_INTERFACE, propName)
		if call.Err != nil {
			log.Println("Error getting property: ", call.Err)
			continue
		}

		value := call.Body[0].(dbus.Variant).Value()
		fmt.Println("Got", value)
		if strings.HasSuffix(propName, "Pixmap") {
			correctedPropName := propName[:len(propName)-6] + "Url"
			dbusValue, ok := value.([][]interface{})
			if !ok {
				log.Println("Expected", propName, "to be of type [][]interface{}, but found", reflect.TypeOf(value))
				continue
			}

			icon := collectPixMap(dbusValue)
			url, err := argb.ServeAsPng(icon)
			if err != nil {
				log.Println("Unable to serve icon as png", err)
				continue
			}
			item.props[correctedPropName] = ".." + url
		} else {
			if strings.HasSuffix(propName, "Name") &&
				item.props["IconThemePath"] != nil &&
				item.props["IconThemePath"] != "" &&
				item.props[propName] != value {
				copyIconDir(item.props["IconThemePath"].(string))
			}
			item.props[propName] = value
		}
	}
}

func (item Item) copy() *Item {
	props := make(map[string]interface{})
	for propName, value := range item.props {
		props[propName] = value // TODO: Maybe not necessary to copy all?
	}
	return &Item{props, item.dbusObj}
}

func MakeItem(dbusObj dbus.BusObject) *Item {
	item := Item{make(map[string]interface{}), dbusObj}
	item.fetchProps(propNames...)
	if iconThemePathProp, ok := item.props["IconThemePath"]; ok {
		if iconThemePath, ok := iconThemePathProp.(string); ok && iconThemePath != "" {
			copyIconDir(iconThemePath)
		}
	}
	return &item
}


func copyIconDir(dir string) {
	if !strings.HasSuffix(dir, "/") {
		dir = dir + "/"
	}

	var destDir = xdg.RuntimeDir + "/org.refude.icon-service-session-icons/"
	var filesCopied = 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error descending into", path)
			return err
		}
		var relPath = path[len(dir):]
		if info.IsDir() {
			if err := os.MkdirAll(destDir+relPath, os.ModePerm); err != nil {
				return err
			}
		} else if _, err := os.Stat(destDir + relPath); err != nil {
			if os.IsNotExist(err) {
				r, err := os.Open(path);
				if err != nil {
					log.Println("Error reading file:", err)
					return err
				}
				defer r.Close()

				w, err := os.Create(destDir + relPath);
				if err != nil {
					log.Println("Error creating file", err)
					return err
				}
				defer w.Close()

				if _, err := io.Copy(w, r); err != nil {
					log.Println("Error copying file", err)
					return err
				}
				filesCopied++
			} else {
				log.Println("Error stat'ing file", err)
				return err
			}
		}
		return nil
	})
	if filesCopied > 0 {
		if _,err := os.Create(destDir + "/marker"); err != nil {
			log.Println("Error updating marker:", err)
		}
	}
}

func StatusNotifierItem(path string, item *Item, signals chan string) {
	service.Map(path, item.copy())
	defer service.Unmap(path)

	for signal := range signals {
		fmt.Println("Received signal:", signal)
		switch (signal) {
		case "NewTitle":
			item.fetchProps("Title")
		case "NewIcon":
			item.fetchProps("IconName", "IconPixmap")
		case "NewAttentionIcon":
			item.fetchProps("AttentionIconName", "AttentionIconPixmap")
		case "NewOverlayIcon":
			item.fetchProps("OverlayIconName", "OverlayIconPixmap")
		case "NewStatus":
			item.fetchProps("Status")
		case "NewIconThemePath":
			item.fetchProps("IconThemePath")
		}
		service.Map(path, item.copy())
	}
	fmt.Println("StatusNotifierItem for", path, "exiting")
}

func collectPixMap(dbusValue [][]interface{}) argb.Icon {
	res := make(argb.Icon, 0)
	for _, arr := range (dbusValue) {
		for len(arr) > 2 {
			width := arr[0].(int32)
			height := arr[1].(int32)
			pixels := arr[2].([]byte)
			res = append(res, argb.Img{Width: width, Height: height, Pixels: pixels})
			arr = arr[3:]
		}
	}

	return res
}
