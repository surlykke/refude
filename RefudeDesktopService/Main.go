// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"reflect"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"
	"github.com/surlykke/RefudeServices/lib"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
)

func serveHttp(w http.ResponseWriter, r *http.Request) {

	var serveResource = func(res resource.Resource) {
		if reflect.ValueOf(res).IsNil() {
			w.WriteHeader(http.StatusNotFound)
		} else if r.Method == "GET" {
			var response = resource.ToJSon(res)
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", response.Etag)
			_, _ = w.Write(response.Data)
		} else if r.Method == "POST" {
			res.POST(w, r)
		} else if r.Method == "PATCH" {
			res.PATCH(w, r)
		} else if r.Method == "DELETE" {
			res.DELETE(w, r)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}

	var serveCollection = func(collection []interface{}) {
		var matcher, err = requests.GetMatcher2(r)
		if err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		} else if matcher != nil {
			var matched = 0
			for i := 0; i < len(collection); i++ {
				if matcher(collection[i]) {
					collection[matched] = collection[i]
					matched++
				}
			}
			collection = collection[0:matched]
		}

		var response = resource.ToJSon(collection)

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", response.Etag)
		_, _ = w.Write(response.Data)
	}

	var path = resource.StandardizedPath(r.URL.Path)
	switch {
	case path == "/windows":
		serveCollection(windows.GetWindows())
	case path.StartsWith("/window/"):
		serveResource(windows.GetWindow(path))
	case path == "/applications":
		serveCollection(applications.GetApplications())
	case path.StartsWith("/application/"):
		serveResource(applications.GetApplication(path))
	case path == "/notifications":
		serveCollection(notifications.GetNotifications())
	case path.StartsWith("/notification/"):
		serveResource(notifications.GetNotification(path))
	case path == "/devices":
		serveCollection(power.GetDevices())
	case path.StartsWith("/device/"):
		serveResource(power.GetDevice(path))
	case path == "/session":
		serveResource(power.Session)
	case path == "/items":
		serveCollection(statusnotifications.GetItems())
	case path.StartsWith("/item/"):
		serveResource(statusnotifications.GetItem(path))
	case path.StartsWith("/itemmenu/"):
		serveResource(statusnotifications.GetMenu(path))
	case path == "/iconthemes":
		serveCollection(icons.GetThemes())
	case path.StartsWith("/icontheme/"):
		serveResource(icons.GetTheme(path))
	case path == "/icons":
		serveCollection(icons.GetIcons())
	case path == "/icon":
		icons.ServeNamedIcon(w, r)
	case path.StartsWith("/icon/"):
		icons.ServeIcon(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	go applications.Run()
	go windows.Run()
	go notifications.Run()
	go power.Run()
	go statusnotifications.Run()
	go icons.Run()

	go lib.Serve("org.refude.desktop-service", http.HandlerFunc(serveHttp))
	_ = http.ListenAndServe(":7938", http.HandlerFunc(serveHttp))
}
