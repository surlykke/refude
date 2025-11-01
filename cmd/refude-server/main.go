// Copyright (c) Christian Surlykke
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"cmp"
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/desktop"
	"github.com/surlykke/refude/internal/desktopactions"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/notifygui"
	"github.com/surlykke/refude/internal/options"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/search"
	"github.com/surlykke/refude/internal/watch"
	"github.com/surlykke/refude/internal/wayland"
	"github.com/surlykke/refude/pkg/bind"
)

func main() {
	var opts = options.GetOpts()

	ServeMap(wayland.WindowMap, "/window/")
	go wayland.Run(opts.IgnoreWinAppIds)

	ServeMap(applications.AppMap, "/application/")
	ServeMap(applications.MimeMap, "/mimetype/")
	go applications.Run()

	if !opts.NoNotifications {
		ServeMap(notifications.NotificationMap, "/notification/")
		go notifications.Run()
		go notifygui.StartGui()
	}

	ServeMap(icons.ThemeMap, "/icontheme/")
	go icons.Run()

	ServeMap(power.DeviceMap, "/device/")
	go power.Run(opts.NoTrayBattery)

	ServeMap(browser.TabMap, "/tab/")
	go browser.Run()

	ServeMap(browser.BookmarkMap, "/bookmark/")
	ServeMap(file.FileMap, "/file/")
	go file.Run()

	ServeMap(desktopactions.PowerActions, "/start/")

	http.Handle("GET /icon", bind.HandlerFunc(icons.GetHandler, bind.Query("name"), bind.QueryOr("size", "32")))
	http.Handle("GET /search", bind.HandlerFunc(search.GetHandler, bind.Query("term")))
	http.Handle("GET /complete", bind.HandlerFunc(completeHandler, bind.Query("prefix")))
	http.Handle("GET /desktop/search", bind.HandlerFunc(desktop.SearchHandler, bind.Query("term")))
	http.Handle("GET /desktop/details", bind.HandlerFunc(desktop.DetailsHandler, bind.Query("path")))

	http.HandleFunc("GET /watch", watch.ServeHTTP)
	go watch.Run()
	http.Handle("GET /desktop/", desktop.StaticServer)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Print("http.ListenAndServe failed:", err)
	}

}

func ServeMap[K cmp.Ordered, V entity.Servable](m *entity.EntityMap[K, V], pathPrefix string) {
	m.SetPrefix(pathPrefix)
	http.Handle("GET "+pathPrefix+"{id...}", bind.HandlerFunc(m.DoGet, bind.Path("id")))
	http.Handle("GET "+pathPrefix+"{$}", bind.HandlerFunc(m.DoGetList))
	http.Handle("POST "+pathPrefix+"{id...}", bind.HandlerFunc(m.DoPost, bind.Path("id"), bind.QueryOr("action", "")))
}

func completeHandler(prefix string) bind.Response {
	var filtered = make([]string, 0, 1000)
	var allPaths = [][]string{
		{"/flash", "/icon?name=", "/desktop/", "/complete?prefix=", "/search?", "/watch"},
		icons.ThemeMap.GetPaths(),
		wayland.WindowMap.GetPaths(),
		applications.AppMap.GetPaths(),
		applications.MimeMap.GetPaths(),
		notifications.NotificationMap.GetPaths(),
		power.DeviceMap.GetPaths(),
		browser.TabMap.GetPaths(),
		browser.BookmarkMap.GetPaths(),
	}

	for _, pathList := range allPaths {
		for _, path := range pathList {
			if strings.HasPrefix(path, prefix) {
				filtered = append(filtered, path)
			}
		}
	}

	return bind.Json(filtered)
}
