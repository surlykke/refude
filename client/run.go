// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package client

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/windows"
)

//go:embed html
var clientResources embed.FS
var StaticServer http.Handler

func init() {
	var tmp http.Handler
	if projectDir, ok := os.LookupEnv("DEV_PROJECT_ROOT_DIR"); ok {
		// Used when developing
		tmp = http.FileServer(http.Dir(projectDir + "/client/html"))
	} else if htmlDir, err := fs.Sub(clientResources, "html"); err == nil {
		// Otherwise, what's baked in
		tmp = http.FileServer(http.FS(htmlDir))
	} else {
		log.Panic(err)
	}
	StaticServer = http.StripPrefix("/refude/html", tmp)
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		if r.URL.Path == "/refude/html/showLauncher" {
			if !windows.RaiseAndFocusNamedWindow("Refude launcher++") {
				xdg.RunCmd("brave-browser", "--app=http://localhost:7938/refude/html/launcher")
			}
			respond.Accepted(w)
			return
		} else if r.URL.Path == "/refude/html/resizeNotifier" {
			var widthS = requests.GetSingleQueryParameter(r, "width", "10")
			var heightS = requests.GetSingleQueryParameter(r, "height", "10")
			if width, err := strconv.Atoi(widthS); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if height, err := strconv.Atoi(heightS); err != nil {
				respond.UnprocessableEntity(w, err)
			} else {
				if width < 10 {
					width = 10
				}

				if height < 10 {
					height = 10
				}

				x, y := calculateNotificationPos(width, height)
				xdg.RunCmd("notifierMove", strconv.Itoa(x), strconv.Itoa(y), strconv.Itoa(width), strconv.Itoa(height))
				respond.Accepted(w)
			}
			return

		}
	}
	StaticServer.ServeHTTP(w, r)
}

func calculateNotificationPos(width, height int) (int, int) {
	// zero values means top right corner
	var corner uint8
	var mX, mY, mW, mH, distX, distY int
outer:
	for _, placement := range config.Notifications.Placement {
		for _, m := range windows.GetMonitors() {
			if placement.Screen == "" && m.Primary || placement.Screen == m.Title {
				mX, mY, mW, mH = m.X, m.Y, m.W, m.H
				corner, distX, distY = placement.Corner, placement.CornerDistX, placement.CornerDistY
				break outer
			}
		}
	}

	switch corner {
	case 0: // top-left
		return mX + distX, mY + distY
	case 1: // top-right
		return mX + mW - width - distX, mY + distY
	case 2: // bottorm right
		return mX + mW - width - distX, mY + mH - height - distY
	case 3: // bottorm left
		return mX + distX, mY + mH - height - distY
	default:
		return 0, 0
	}

}
