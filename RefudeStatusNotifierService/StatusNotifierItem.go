package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/argb"
	"regexp"
	"errors"
	"github.com/godbus/dbus"
	"reflect"
	"github.com/surlykke/RefudeServices/lib/common"
)

type Item map[string]interface{}


type ToolTip struct {
	Title       string
	Description string
	IconName    string
	IconUrl     string
}

func (item Item) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, item)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}



func StatusNotifierItem(serviceName string, signals chan string) {

	var item = make(Item)

	fetchProperties(serviceName, item, itemFields...)
	itemId, ok := item["Id"]
	if !ok || itemId == "" {
		fmt.Println("No Id on ", serviceName)
		return
	}
	path := "/items/" + itemId.(string)
	serveCopy(path, item)

	defer service.Unmap(path)

	for signal := range signals {
		fmt.Println("Got signal", signal)
		switch (signal) {
		case "NewTitle":
			fetchProperties(serviceName, item, "Title")
		case "NewIcon":
			fetchProperties(serviceName, item, "IconName", "IconPixmap")
		case "NewAttentionIcon":
			fetchProperties(serviceName, item, "AttentionIconName", "AttentionIconPixmap")
		case "NewOverlayIcon":
			fetchProperties(serviceName, item, "OverlayIconName", "OverlayIconPixmap")
		case "NewStatus":
			fetchProperties(serviceName, item, "Status")
		}

		serveCopy(path, item)
	}
}

var serviceNameReg = regexp.MustCompile(`org.kde.StatusNotifierItem-(.*)`)

func getId(serviceName string) (string, error) {
	m := serviceNameReg.FindStringSubmatch(serviceName)
	if len(m) > 0 {
		return  m[1], nil
	} else {
		return "", errors.New(serviceName + " does not match")
	}
}

var itemFields = []string{
			"Id",
			"Category",
			"Status",
			"Title",
			"ItemIsMenu",
			"IconName",
			"AttentionIconName",
			"OverlayIconName",
			"AttentionMovieName",
			"IconPixmap",
			"AttentionIconPixmap",
			"OverlayIconPixmap",
//			"ToolTip",
}

func fetchProperties(serviceName string, dest Item, propNames...string) {
	obj := conn.Object(serviceName, dbus.ObjectPath(ITEM_PATH))
	for _,propName := range propNames {
		call := obj.Call("org.freedesktop.DBus.Properties.Get", dbus.Flags(0), ITEM_INTERFACE, propName)
		if call.Err != nil {
			fmt.Println("Error getting", propName, call.Err)
			continue
		}
		value := call.Body[0].(dbus.Variant).Value()
		switch value.(type) {
		case bool, string:
			dest[propName] = value
		case [][]interface{}:
			icon := collectPixMap(value.([][]interface{}))
			url, err := argb.ServeAsPng(icon)
			if err == nil {
				dest[propName] = url
			}
		default:
			fmt.Println("Unable to handle ", reflect.TypeOf(value))
		}
	}
}

func collectPixMap(dbusValue [][]interface{}) argb.Icon {
	res := make(argb.Icon, 0)
	for _,arr := range(dbusValue) {
		for len (arr) > 2 {
			width := arr[0].(int32)
			height := arr[1].(int32)
			pixels := arr[2].([]byte)
			res = append(res, argb.Img{Width: width, Height: height, Pixels: pixels})
			arr = arr[3:]
		}
	}

	return res
}

func serveCopy(path string, m Item) {
	copy := make(Item)
	for key,value := range m {
		copy[key] = value
	}

	service.Map(path, copy)
}

