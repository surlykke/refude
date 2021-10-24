package client

import (
	"embed"
	"io/fs"
	"net/http"
	"os/exec"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/respond"
)

//go:embed html
var clientResources embed.FS
var StaticServer http.Handler

func init() {
	// When Installed
	if htmlDir, err := fs.Sub(clientResources, "html"); err != nil {
		log.Panic(err)
	} else {
		StaticServer = http.FileServer(http.FS(htmlDir))
	}

	//When developing
	//StaticServer = http.FileServer(http.Dir("/home/surlykke/RefudeServices/client/html"))
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
