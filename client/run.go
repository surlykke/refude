// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package client

import (
	"embed"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"strconv"

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
			if !windows.WM.RaiseAndFocusNamedWindow("Refude launcher++") {
				xdg.RunCmd("brave-browser", "--app=http://localhost:7938/refude/html/launcher")
			}
			respond.Accepted(w)
			return
		} else if r.URL.Path == "/refude/html/resizeNotifier" {
			var widthS = requests.GetSingleQueryParameter(r, "width", "10")
			var heightS = requests.GetSingleQueryParameter(r, "height", "10")
			if width, err := strconv.Atoi(widthS); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if width < 2 {
				respond.UnprocessableEntity(w, errors.New("width and height must each be > 2")) 
			} else if height, err := strconv.Atoi(heightS); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if height < 2 {
				respond.UnprocessableEntity(w, errors.New("width and height must each be > 2")) 
			} else if ! windows.WM.ResizeNamedWindow("Refude notifier", uint32(width), uint32(height)) {
				respond.NotFound(w)
			} else {
				respond.Accepted(w)
			}
			return  
		}

	}
	StaticServer.ServeHTTP(w, r)
}
