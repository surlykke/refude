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
	"sync"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/pubsub"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/wayland"
	"golang.org/x/net/websocket"
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
	fmt.Println("client.Run:", r.URL.Path)
	if r.URL.Path == "/refude/html/showlauncher" {
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			showLauncher()
			respond.Accepted(w)
		}
	} else if r.URL.Path == "/refude/html/hidelauncher" {
		if r.Method != "POST" {
			respond.NotAllowed(w)
		} else {
			var restore = r.URL.Query()["restore"]
			if slice.Contains(restore, "tab") {
				commands.Publish(command{"restoreTab"})
			}
			if slice.Contains(restore, "window") {
				wayland.ActivateRememberedActive()
			}

			commands.Publish(command{"hide"})
			respond.Accepted(w)
		}
	} else {
		StaticServer.ServeHTTP(w, r)
	}
}

func showLauncher() {
	nclientsLock.Lock()
	var haveClient = nclients > 0
	nclientsLock.Unlock()
	wayland.RememberActive()
	if haveClient {
		commands.Publish(command{"show"})
	} else {
		xdg.RunCmd("xdg-open", "http://localhost:7938/refude/html/launcher/")
	}

}

type command struct {
	Command string
}

var commands = pubsub.MakePublisher[command]()
var nclients int = 0
var nclientsLock sync.Mutex

var CommandSocketHandler = websocket.Handler(func(conn *websocket.Conn) {
	nclientsLock.Lock()
	nclients++
	nclientsLock.Unlock()
	var sub = commands.Subscribe()
	for {
		var command = sub.Next()
		if err := websocket.JSON.Send(conn, command); err != nil {
			log.Warn(err)
			nclientsLock.Lock()
			nclients--
			nclientsLock.Unlock()
			return
		}
	}
})
