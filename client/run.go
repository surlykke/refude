package client

import (
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/exec"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/respond"
)

//go:embed html
var clientResources embed.FS
var StaticServer http.Handler

func init() {
	if projectDir, ok := os.LookupEnv("DEV_PROJECT_ROOT_DIR"); ok {
		// Used when developing
		StaticServer = http.FileServer(http.Dir(projectDir + "/client/html"))
	} else if htmlDir, err := fs.Sub(clientResources, "html"); err == nil {
		// Otherwise, what's baked in
		StaticServer = http.FileServer(http.FS(htmlDir))
	} else {
		log.Panic(err)
	}
}

var events = make(chan string)

func Run() {
	for ev := range events {
		switch ev {
		case "dismiss":
			if err := exec.Command("hideRefudeClient").Run(); err != nil {
				log.Warn("Error hiding client:", err)
			}
		}
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/client/dismiss":
		if r.Method == "POST" {
			events <- "dismiss"
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	default:
		respond.NotFound(w)
	}
}
