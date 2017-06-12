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


var propNames = []string{
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

type Item struct {
	props   map[string]interface{}
	dbusObj dbus.BusObject
}

func (item Item) fetchProps(propNames...string) {
	for _,propName := range propNames {
		delete(item.props, propName)

		fmt.Println("Fetch", propName)
		call := item.dbusObj.Call(PROPERTIES_INTERFACE + ".Get", dbus.Flags(0), ITEM_INTERFACE, propName)
		if call.Err != nil {
			log.Println(call.Err)
			continue
		}

		value := call.Body[0].(dbus.Variant).Value()
		if strings.HasSuffix(propName, "Pixmap") {
			correctedPropName := propName[:len(propName)-6] + "Url"
			dbusValue, ok := value.([][]interface{})
			if !ok {
				log.Println("Expected", propName, "to be of type [][]interface{}, but found", reflect.TypeOf(value))
				continue
			}

			icon := collectPixMap(dbusValue)
			url, err := argb.ServeAsPng(icon)
			if  err != nil {
				log.Println("Unable to serve icon as png", err)
				continue
			}
			item.props[correctedPropName] = ".." + url
		} else {
			item.props[propName] = value
		}
	}
}

func (item Item) copy() Item {
	props := make(map[string]interface{})
	for propName, value := range item.props {
		props[propName] = value  // TODO: Maybe not necessary to copy all?
	}
	return Item{props, item.dbusObj}
}

func MakeItem(dbusObj dbus.BusObject) Item {
	item := Item{make(map[string]interface{}), dbusObj}
	item.fetchProps(propNames...)
	return item
}

func (item Item) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, item.props)
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

		item.dbusObj.Call("org.kde.StatusNotifierItem." + method, dbus.Flags(0), x, y)
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}



func StatusNotifierItem(serviceOwner string, objectPath dbus.ObjectPath, signals chan string) {

	item := MakeItem(conn.Object(serviceOwner, objectPath))

	path := "/items/" + serviceOwner[1:] // Omit leading colon
	service.Map(path, item.copy())

	defer service.Unmap(path)

	for signal := range signals {
		switch (signal) {
		case "NewTitle": item.fetchProps("Title")
		case "NewIcon": item.fetchProps("IconName", "IconPixmap")
		case "NewAttentionIcon": item.fetchProps("AttentionIconName", "AttentionIconPixmap")
		case "NewOverlayIcon": item.fetchProps("OverlayIconName", "OverlayIconPixmap")
		case "NewStatus": item.fetchProps("Status")
		}
		service.Map(path, item.copy())
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


