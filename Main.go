// Copyright (c) Christian Surlykke
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/browser"
	"github.com/surlykke/RefudeServices/desktop"
	"github.com/surlykke/RefudeServices/desktopactions"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/bind"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/options"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/search"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/wayland"

	_ "net/http/pprof"
)

func main() {
	var opts = options.GetOpts()
	bind.ServeMap("/window/", wayland.WindowMap)
	bind.ServeMap("/application/", applications.AppMap)
	bind.ServeMap("/mimetype/", applications.MimeMap)
	bind.ServeMap("/notification/", notifications.NotificationMap)
	bind.ServeMap("/icontheme/", icons.ThemeMap)
	bind.ServeMap("/device/", power.DeviceMap)
	bind.ServeMap("/tab/", browser.TabMap)
	bind.ServeMap("/bookmark/", browser.BookmarkMap)
	bind.ServeMap("/file/", file.FileMap)

	bind.ServeFunc("GET /icon", icons.GetHandler, `query:"name"`, `query:"size,default=32"`)
	bind.ServeFunc("GET /search", search.GetHandler, `query:"term"`)
	bind.ServeFunc("GET /start", desktopactions.GetHandler)
	bind.ServeFunc("POST /start", desktopactions.PostHandler, `query:"action"`)
	bind.ServeFunc("GET /flash", notifications.FlashHandler)
	//	bind.ServeFunc("POST /tabsink", browser.TabsDoPost, `query:"browserName,required"`, `body:"json"`)
	//  bind.ServeFunc("POST /bookmarksink", browser.BookmarksDoPost, `body:"json"`)
	bind.ServeFunc("GET /complete", completeHandler, `query:"prefix"`)
	bind.ServeFunc("GET /desktop/search", desktop.SearchHandler, `query:"term"`)

	http.HandleFunc("GET /watch", watch.ServeHTTP)
	http.Handle("GET /desktop/", desktop.StaticServer)
	http.HandleFunc("GET /browser/socket", browser.ServeHTTP)
	go icons.Run()
	go wayland.Run(opts.IgnoreWinAppIds)
	go applications.Run()
	if !opts.NoNotifications {
		go notifications.Run()
	}
	go power.Run()
	go file.Run()
	go browser.Run()

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}
}

type pref struct {
	prefix string
}

func completeHandler(prefix string) response.Response {
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

	return response.Json(filtered)
}
