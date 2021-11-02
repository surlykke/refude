package client

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
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

var events = make(chan string)

func Run() {
	for ev := range events {
		switch ev {
		case "dismiss":
			if err := exec.Command("browserHide").Run(); err != nil {
				log.Warn("Error hiding client:", err)
			}
		}
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/refude/html/") {
		StaticServer.ServeHTTP(w, r)
	} else if r.URL.Path == "/refude/browser/dismiss" {
		if r.Method == "POST" {
			events <- "dismiss"
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	} else if r.URL.Path == "/refude/panel/resize" {
		if r.Method == "POST" {
			if width, err := strconv.ParseUint(requests.GetSingleQueryParameter(r, "width", ""), 10, 32); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if height, err := strconv.ParseUint(requests.GetSingleQueryParameter(r, "height", ""), 10, 32); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if !windows.ResizeNamedWindow("org.refude.panel", uint32(width), uint32(height)) {
				respond.NotFound(w)
			} else {
				respond.Accepted(w)
			}
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}
