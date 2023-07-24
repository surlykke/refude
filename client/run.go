// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package client

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"

	"github.com/surlykke/RefudeServices/lib/log"
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
	fmt.Println("client.ServeHTTP, path:", r.URL.Path, ", query:", r.URL.Query())
	if r.URL.Path == "/refude/html/showlauncher" {
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			Show("launcher", true)
			respond.Accepted(w)
		} 	
	} else {
		StaticServer.ServeHTTP(w, r)
	}
}


func Show(app string, focus bool)  {
	var appName = "Refude " + app
	if ! windows.PurgeAndShow(appName, focus) {
		xdg.RunCmd(xdg.BrowserCommand, "--app=http://localhost:7938/refude/html/" + app)
	}
}
		


