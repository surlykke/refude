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
	fmt.Println("client.ServeHTTP, path:", r.URL.Path, ", query:", r.URL.Query())
	if slice.Among(r.URL.Path, "/refude/html/show", "/refude/html/hide") {
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else if app := requests.GetSingleQueryParameter(r, "app", ""); app != "launcher" && app != "notifier" {
			respond.UnprocessableEntity(w, errors.New("'app' must be given and one of 'launcher' or 'notifier'"))
		} else if r.URL.Path == "/refude/html/hide" {
			windows.PurgeAndHide("Refude " + app)
			respond.Accepted(w)
		} else if app == "launcher" {
			x, y, width, height := calculatePosAndSize(config.Launcher.Placement, 300, 400)
			if (!windows.MoveAndResize("Refude launcher", int32(x), int32(y), uint32(width), uint32(height))) {
				xdg.RunCmd(xdg.BrowserCommand, "--app=http://localhost:7938/refude/html/launcher/")
			}
			respond.Accepted(w)
		} else if app == "notifier" {
			if width, height, err := getWidthAndHeight(r); err != nil {
				respond.UnprocessableEntity(w, err)
			}  else {
				var x, y, width, height = calculatePosAndSize(config.Notifications.Placement, width, height)
				if (!windows.MoveAndResize("Refude notifier", int32(x), int32(y), uint32(width), uint32(height))) {
					xdg.RunCmd(xdg.BrowserCommand, "--app=http://localhost:7938/refude/html/notifier/")
				}		
				respond.Accepted(w)
			}
		}
	} else {
		StaticServer.ServeHTTP(w, r)
	}
}

func getWidthAndHeight(r *http.Request) (uint, uint, error) {
	if width, err := requests.GetPosInt(r, "width"); err != nil {
		return 0, 0, err
	} else if height, err := requests.GetPosInt(r, "height"); err != nil {
		return 0, 0, err
	} else {
		return width, height, nil
	}
}


func calculatePosAndSize(placementList []config.Placement, widthHint uint, heightHint uint) (uint, uint, uint, uint) {
	// zero values means top right corner
	var corner uint8
	var mX, mY, mW, mH, relX, relY, width, height uint
	var mdList = monitor.GetMonitors()
	var tmp = append(placementList, config.Placement{}) // The appended assures we'll always get a match below

outer:
	for _, placement := range tmp {
		for _, m := range mdList {
			if placement.Screen == "" && m.Primary || placement.Screen == m.Title {
				mX, mY, mW, mH = uint(m.X), uint(m.Y), uint(m.W), uint(m.H)
				corner, relX, relY, width, height = placement.Corner, placement.CornerdistX, placement.CornerdistY, placement.Width, placement.Height
				break outer
			}
		}
	}

	if width == 0 {
		width = widthHint
	}
	if height == 0 {
		height = heightHint
	}

	var x,y uint
	switch corner {
	case 0: // top-left
		x,y = mX + relX, mY + relY
	case 1: // top-right
		x,y = mX + mW - width - relX - 2, mY + relY
	case 2: // bottorm right
		x,y = mX + mW - width - relX - 2, mY + mH - height - relY - 2
	case 3: // bottorm left
		x,y = mX + relX, mY + mH - height - relY - 2
	default:
		x,y = 0, 0
	}
	return x,y,width,height
}

