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
	"github.com/surlykke/refude/internal/lib/bind"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/options"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/search"
	"github.com/surlykke/refude/internal/watch"
	"github.com/surlykke/refude/internal/wayland"
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
	}

	ServeMap(icons.ThemeMap, "/icontheme/")
	go icons.Run()

	ServeMap(power.DeviceMap, "/device/")
	go power.Run()

	ServeMap(browser.TabMap, "/tab/")
	go browser.Run()

	ServeMap(browser.BookmarkMap, "/bookmark/")
	ServeMap(file.FileMap, "/file/")
	go file.Run()

	ServeMap(desktopactions.PowerActions, "/start/")

	http.HandleFunc("GET /icon", bind.ServeFunc(icons.GetHandler, bind.Query("name"), bind.QueryOpt("size", "32")))
	http.HandleFunc("GET /search", bind.ServeFunc(search.GetHandler, bind.Query("term")))
	http.HandleFunc("GET /flash", bind.ServeFunc(notifications.FlashHandler))
	http.HandleFunc("GET /complete", bind.ServeFunc(completeHandler, bind.Query("prefix")))
	http.HandleFunc("GET /desktop/search", bind.ServeFunc(desktop.SearchHandler, bind.Query("term")))
	http.HandleFunc("GET /desktop/details", bind.ServeFunc(desktop.DetailsHandler, bind.Query("path")))

	http.HandleFunc("GET /watch", watch.ServeHTTP)
	http.Handle("GET /desktop/", desktop.StaticServer)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Print("http.ListenAndServe failed:", err)
	}

}

func ServeMap[K cmp.Ordered, V entity.Servable](em *entity.EntityMap[K, V], pathPrefix string) {
	em.SetPrefix(pathPrefix)
	http.HandleFunc("GET "+pathPrefix+"{id...}", bind.ServeFunc(em.DoGetSingle, bind.Path("id")))
	http.HandleFunc("GET "+pathPrefix+"{$}", bind.ServeFunc(em.DoGetAll))
	http.HandleFunc("POST "+pathPrefix+"{id...}", bind.ServeFunc(em.DoPost, bind.Path("id"), bind.QueryOpt("action", "")))
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
