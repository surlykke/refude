package browsertabs

import (
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"golang.org/x/net/websocket"
)

var connections = make(map[*websocket.Conn]bool, 10)
var connectionsLock sync.Mutex



type Tab struct {
	resource.BaseResource
}

func (this *Tab) DoPost(w http.ResponseWriter, r *http.Request) {
	fmt.Println("DoPost...")
	connectionsLock.Lock()
	for conn := range connections {
		fmt.Println("Sending", this.Id)
		websocket.JSON.Send(conn, this.Id)
	}
	connectionsLock.Unlock()
	respond.Accepted(w)
}

func (this *Tab) RelevantForSearch() bool {
	return ! strings.HasPrefix(this.Title, "Refude launcher")
}

var Tabs = resource.MakeCollection[*Tab]("/tab/")


var WebsocketHandler = websocket.Handler(func(conn *websocket.Conn) {
	connectionsLock.Lock()
	connections[conn] = true
	connectionsLock.Unlock()

	defer fmt.Println("receiver done")
	fmt.Println("Start receiving")
	for {
		var data = make([]map[string]string, 30)
		if err := websocket.JSON.Receive(conn, &data); err != nil {
			log.Warn(err)
			connectionsLock.Lock()
			delete(connections, conn)
			connectionsLock.Unlock()
			conn.Close()
			return
		} else {
			var tabs = make([]*Tab, 0, len(data))
			for _, d := range data {
				tabs = append(tabs, &Tab{
					BaseResource: resource.BaseResource{
						Id:      d["id"],
						Title:   d["title"],
						Comment: d["url"],
						Profile: "browsertab",
					}})
			}
			Tabs.ReplaceWith(tabs)
		}
	}
})
