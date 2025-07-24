// Copyright (c) Christian Surlykke
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"net/http"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/desktop"
	"github.com/surlykke/refude/internal/desktopactions"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/bind"
	"github.com/surlykke/refude/internal/lib/log"
	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/options"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/search"
	"github.com/surlykke/refude/internal/watch"
	"github.com/surlykke/refude/internal/wayland"
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
	bind.ServeMap("/start/", desktopactions.PowerActions)

	bind.ServeFunc("GET /icon", icons.GetHandler, `query:"name"`, `query:"size,default=32"`)
	bind.ServeFunc("GET /search", search.GetHandler, `query:"term"`)
	bind.ServeFunc("GET /flash", notifications.FlashHandler)
	bind.ServeFunc("GET /complete", completeHandler, `query:"prefix"`)
	bind.ServeFunc("GET /desktop/search", desktop.SearchHandler, `query:"term"`)
	bind.ServeFunc("GET /desktop/details", desktop.DetailsHandler, `query:"path"`)

	http.HandleFunc("GET /watch", watch.ServeHTTP)
	http.Handle("GET /desktop/", desktop.StaticServer)
	go icons.Run()
	go wayland.Run(opts.IgnoreWinAppIds)
	go applications.Run()
	go browser.Run()
	if !opts.NoNotifications {
		go notifications.Run()
	}
	go power.Run()
	go file.Run()

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Warn("http.ListenAndServe failed:", err)
	}

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
