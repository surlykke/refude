// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package client

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/surlykke/RefudeServices/config"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/monitor"
	"github.com/surlykke/RefudeServices/x11"
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
	if slice.Among(r.URL.Path, "/refude/html/show", "/refude/html/hide", "/refude/html/resize") {
		if r.Method != "POST" {
			respond.NotAllowed(w)
			return
		} else if app := requests.GetSingleQueryParameter(r, "app", ""); app != "launcher" && app != "notifier" {
			respond.UnprocessableEntity(w, errors.New("'app' must be given"))
		} else if r.URL.Path == "/refude/html/show" {
			if !x11.PurgeAndShow("localhost__refude_html_"+app, app == "launcher") {
				xdg.RunCmd(xdg.BrowserCommand, fmt.Sprintf("--app=http://localhost:7938/refude/html/%s/", app))
			}
			respond.Accepted(w)
		} else if r.URL.Path == "/refude/html/hide" {
			x11.PurgeAndHide("localhost__refude_html_" + app)
			respond.Accepted(w)
		} else if width, ok := requests.GetInt(r, "width"); !ok || width <= 0 {
			respond.UnprocessableEntity(w, errors.New("'width' must be given, and be a positive int"))
		} else if height, ok := requests.GetInt(r, "height"); !ok || height <= 0 {
			respond.UnprocessableEntity(w, errors.New("'height' must be given, and be a positive int"))
		} else {
			var x, y int
			if app == "launcher" {
				x, y = calculateLauncherPos(width, height)
			} else {
				x, y = calculateNotifierPos(width, height)
			}
			x11.MoveAndResize("localhost__refude_html_"+app, int32(x), int32(y), uint32(width), uint32(height))
			respond.Accepted(w)
		}
	}
	StaticServer.ServeHTTP(w, r)
}

func calculateLauncherPos(width, height int) (int, int) {
	var x, y int = 0, 0
	var mdList = monitor.GetMonitors()

	for _, m := range mdList {
		if m.Primary {
			x = m.X + (m.W-width)/2
			y = m.Y + (m.H-height)/2
			break
		}
	}

	return x, y
}

func calculateNotifierPos(width, height int) (int, int) {
	// zero values means top right corner
	var corner uint8
	var mX, mY, mW, mH, distX, distY int
	var mdList = monitor.GetMonitors()
	var tmp = append(config.Notifications.Placement, config.Placement{}) // The appended assures we'll always get a match below

outer:
	for _, placement := range tmp {
		for _, m := range mdList {
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
		return mX + mW - width - distX - 2, mY + distY
	case 2: // bottorm right
		return mX + mW - width - distX - 2, mY + mH - height - distY - 2
	case 3: // bottorm left
		return mX + distX, mY + mH - height - distY - 2
	default:
		return 0, 0
	}

}
