// Copyright (c) Christian Surlykke
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/browser"
	"github.com/surlykke/refude/internal/desktop/pwa"
	"github.com/surlykke/refude/internal/desktopactions"
	"github.com/surlykke/refude/internal/file"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/notifications"
	"github.com/surlykke/refude/internal/notifygui"
	"github.com/surlykke/refude/internal/options"
	"github.com/surlykke/refude/internal/power"
	"github.com/surlykke/refude/internal/rawfile"
	"github.com/surlykke/refude/internal/search"
	"github.com/surlykke/refude/internal/wayland"
)

func main() {
	var opts = options.GetOpts()

	go wayland.Run(opts.IgnoreWinAppIds)
	go applications.Run()
	go notifications.Run(opts.NoNotifications)
	go notifygui.StartGui(opts.NoNotifications)
	go desktopactions.Run()
	go icons.Run()
	go power.Run(opts.NoTrayBattery)
	go browser.Run()
	go file.Run()
	go rawfile.Run()
	go pwa.Run()
	go search.Run()
	//	go network.Run()

	http.HandleFunc("GET /complete", CompleteHandler)
	http.HandleFunc("/", resource.ServeHTTP)

	if err := http.ListenAndServe(":7938", nil); err != nil {
		log.Print("http.ListenAndServe failed:", err)
	}

}

func CompleteHandler(w http.ResponseWriter, r *http.Request) {
	var filtered = make([]string, 0, 1000)
	var prefix = utils.QueryParam(r, "prefix")
	for _, path := range resource.GetPaths() {
		if strings.HasPrefix(path, prefix) {
			filtered = append(filtered, path)
		}
	}

	for _, path := range []string{"/flash", "/icon?name=", "/desktop/", "/complete?prefix=", "/search?", "/watch"} {
		if strings.HasPrefix(path, prefix) {
			filtered = append(filtered, path)
		}
	}

	respond.AsJson(w, filtered)
}
