package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/service"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/argb"
	"github.com/godbus/dbus"
	"reflect"
	"github.com/surlykke/RefudeServices/lib/common"
	"strings"
	"log"
	"strconv"
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
	} else if r.Method == "POST" {
		method := common.GetSingleQueryParameter(r, "method", "Activate")
		x, errX := strconv.Atoi(common.GetSingleQueryParameter(r, "x", "0"))
		y, errY := strconv.Atoi(common.GetSingleQueryParameter(r, "y", "0"))
		if (method != "Activate" && method != "SecondaryActivate" && method != "ContextMenu") ||
			errX != nil ||
			errY != nil {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		}

        obj := conn.Object(item["serviceName"].(string), "/StatusNotifierItem")
		obj.Call("org.kde.StatusNotifierItem." + method, dbus.Flags(0), x, y)
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
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

func StatusNotifierItem(serviceName string, signals chan string) {

	var item = make(Item)
	item["serviceName"] = serviceName
	fetchProperties(serviceName, item, itemFields...)
	itemId, ok := item["Id"]
	if !ok || itemId == "" {
		fmt.Println("No Id on ", serviceName)
		return
	}
	path := "/items/" + serviceName[len("org.kde.StatusNotifierItem-"):]
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


func fetchProperties(serviceName string, dest Item, propNames...string) {
	obj := conn.Object(serviceName, dbus.ObjectPath(ITEM_PATH))
	for _,propName := range propNames {
		call := obj.Call("org.freedesktop.DBus.Properties.Get", dbus.Flags(0), ITEM_INTERFACE, propName)
		if call.Err != nil {
			fmt.Println("Error getting", propName, call.Err)
			continue
		}
		value := call.Body[0].(dbus.Variant).Value()
		switch {
		case strings.HasSuffix(propName, "Pixmap"):
			if dbusValue, ok := value.([][]interface{}); ok {
				icon := collectPixMap(dbusValue)
				if url, err := argb.ServeAsPng(icon); err == nil {
					url = ".." + url
					dest[strings.Replace(propName, "Pixmap", "Url", -1)] = url
				} else {
					log.Println(err)
				}
			} else {
				log.Println("Expected", propName, "to be of type [][]interface{}, but found", reflect.TypeOf(value))
			}
		case "ItemIsMenu" == propName:
		if dbusValue, ok := value.(bool); ok {
				dest[propName] = dbusValue
			} else {
				log.Println("Expected", propName, "to be of type bool, but found", reflect.TypeOf(value))
			}
		default:
			if dbusValue, ok := value.(string); ok {
				dest[propName] = dbusValue
			} else {
				log.Println("Expected", propName, "to be of type string, but found", reflect.TypeOf(value))
			}
		}

		if strings.HasSuffix(propName, "Pixmap") {




		}
		switch value.(type) {
		case bool, string:
			dest[propName] = value
		case [][]interface{}:

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

