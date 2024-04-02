// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package desktop

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"os"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/watch"
	"github.com/surlykke/RefudeServices/wayland"
)

//go:embed html
var sources embed.FS

var mainTemplate *template.Template
var StaticServer http.Handler


func init() {
	var bytes []byte
	var err error

	if bytes, err = sources.ReadFile("html/mainTemplate.html"); err != nil {
		log.Panic(err)
	} 

	mainTemplate = template.Must(template.New("mainTemplate").Parse(string(bytes)))
}

func init() {
	var tmp http.Handler

	if projectDir, ok := os.LookupEnv("DEV_PROJECT_ROOT_DIR"); ok {
		// Used when developing
		tmp = http.FileServer(http.Dir(projectDir + "/desktop/html"))
	} else if htmlDir, err := fs.Sub(sources, "html"); err == nil {
		// Otherwise, what's baked in
		tmp = http.FileServer(http.FS(htmlDir))
	} else {
		log.Panic(err)
	}
	StaticServer = http.StripPrefix("/desktop", tmp)

}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var resourcePath = requests.GetSingleQueryParameter(r, "resource", "/start")
	var term = requests.GetSingleQueryParameter(r, "search", "")
	switch(r.URL.Path) {


	case "/desktop/", "/desktop/index.html":
		if m, ok := fetchResourceData(resourcePath, term); ok {
			if err := mainTemplate.Execute(w, m); err != nil {
				log.Warn("Error executing mainTemplate:", err)
			}
		} else {
			respond.NotFound(w)
		}
	case "/desktop/show":
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			wayland.RememberActive()
			watch.Publish("showDesktop", "")			
			respond.Accepted(w)
		}
	case "/desktop/hide":
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			var restore = r.URL.Query()["restore"]
			if slice.Contains(restore, "window") {
				wayland.ActivateRememberedActive()
			}
			watch.Publish("hideDesktop", "")
			respond.Accepted(w)
		}
	case "mainTemplate.html":
		respond.NotFound(w)
	default:
		StaticServer.ServeHTTP(w, r)
	}
}
