// Copyright (c) Christian Surlykke
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/desktop"
	"github.com/surlykke/refude/internal/desktopactions"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/network"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/notifygui"
	"github.com/surlykke/refude/internal/options"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/search"
	"github.com/surlykke/refude/internal/watch"
	"github.com/surlykke/refude/internal/wayland"
)

func main() {
	var opts = options.GetOpts()

	go wayland.Run(opts.IgnoreWinAppIds)
	go applications.Run()
	go notifications.Run(opts.NoNotifications)
	go notifygui.StartGui(opts.NoNotifications)

	go icons.Run()
	go power.Run(opts.NoTrayBattery)
	go browser.Run()
	go file.Run()
	go desktopactions.Run()
	go desktop.Run()
	go search.Run()
	go network.Run()
	go watch.Run()

	http.HandleFunc("GET /complete", CompleteHandler)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Print("http.ListenAndServe failed:", err)
	}

}

func CompleteHandler(w http.ResponseWriter, r *http.Request) {
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
	var prefix = utils.QueryParam(r, "prefix")
	for _, pathList := range allPaths {
		for _, path := range pathList {
			if strings.HasPrefix(path, prefix) {
				filtered = append(filtered, path)
			}
		}
	}

	respond.AsJson(w, filtered)
}
