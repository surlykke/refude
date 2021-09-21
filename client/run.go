package client

import (
	"embed"
	"fmt"
	"net/http"
	"os/exec"

	"github.com/gorilla/websocket"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

//go:embed refude
var clientResources embed.FS
var StaticServer = http.FileServer(http.Dir("/home/surlykke/RefudeServices/client"))

//http.FileServer(http.FS(clientResources))

type eventType int

const (
	command eventType = iota
	controllerConnected
	controllerDisconnected
)

var typeText = map[eventType]string{
	command:                "command",
	controllerConnected:    "controllerConnected",
	controllerDisconnected: "controllerDisconnected",
}

func (et eventType) String() string {
	return typeText[et]
}

type event struct {
	t    eventType
	data interface{}
}

var events = make(chan event)

func Run() {
	var controlConn *websocket.Conn

	for ev := range events {
		fmt.Println("event: ", ev)
		switch ev.t {
		case command:
			if controlConn != nil {
				if err := controlConn.WriteMessage(websocket.TextMessage, []byte(ev.data.(string))); err != nil {
					log.Warn(err)
				}
			} else {
				if err := exec.Command("ensureBrowserRunning.sh").Run(); err != nil {
					log.Warn("Error launching browser:", err)
				}
			}
		case controllerConnected:
			controlConn = ev.data.(*websocket.Conn)
		case controllerDisconnected:
			controlConn = nil
		}
	}
}

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Request:", r.Method, r.URL.Path)
	switch r.URL.Path {
	case "/client":
		if r.Method == "POST" {
			var direction = requests.GetSingleQueryParameter(r, "direction", "down")
			events <- event{command, direction}
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	case "/client/control":
		if r.Method == "GET" {
			if c, err := upgrader.Upgrade(w, r, nil); err != nil {
				fmt.Println(err)
			} else {
				events <- event{controllerConnected, c}
			}
		} else {
			respond.NotAllowed(w)
		}
	}
}
