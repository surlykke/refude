package browser

import (
	"context"
	"net/http"
	"net/url"

	"github.com/aquilax/truncate"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/repo"
)

var TabMap = repo.MakeSynkMap[string, *Tab]()
var BookmarkMap = repo.MakeSynkMap[string, *Bookmark]()

type message = struct {
	BrowserName string
	MsgType     string
	Data        []map[string]string
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if con, err := websocket.Accept(w, r, &websocket.AcceptOptions{OriginPatterns: []string{"lcnmbmoiobgochkfoenopkgnoojgbeio"}}); err != nil {
		log.Error("Accept error:", err)
		return
	} else {
		defer con.CloseNow()
		socketsChan <- con
		var ctx = context.Background()
		con.SetReadLimit(1048576)
		for {
			var msg = &message{}
			if err := wsjson.Read(ctx, con, &msg); err != nil {
				log.Error("Error during read:", err)
				break
			}
			if msg.MsgType == "tabs" {
				var mapOfTabs = make(map[string]*Tab, len(msg.Data))
				for _, d := range msg.Data {
					var tab = &Tab{
						Base:      *entity.MakeBase(truncate.CutEllipsisStrategy{}.Truncate(d["title"], 60), msg.BrowserName+" tab", icon.Name(d["favicon"]), mediatype.Tab),
						Id:        d["id"],
						BrowserId: msg.BrowserName,
						Url:       d["url"]}
					tab.AddAction("", "Focus", "")
					//tab.AddDeleteAction("close", title, "Close tab", "")

					mapOfTabs[tab.Id] = tab
				}
				TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == msg.BrowserName })
			} else if msg.MsgType == "bookmarks" {
				var mapOfBookmarks = make(map[string]*Bookmark, len(msg.Data))
				for _, d := range msg.Data {
					if id, title, bmUrl := d["id"], d["title"], d["url"]; bmUrl == "" {
						continue
					} else {

						var iconUrl = icon.Name("https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=" + url.QueryEscape(bmUrl))
						var bookMark = Bookmark{
							Base:        *entity.MakeBase(title, "", iconUrl, mediatype.Bookmark),
							Id:          id,
							ExternalUrl: bmUrl}
						bookMark.AddAction("", "Open", "")
						mapOfBookmarks[bookMark.Id] = &bookMark
					}

				}
				BookmarkMap.ReplaceAll(mapOfBookmarks)
			}
		}

	}
}

type tabCommand struct {
	BrowserName string `json:"browserName"`
	Operation   string `json:"operation"`
	TabId       string `json:"tabId"`
}

var commands = make(chan tabCommand)
var socketsChan = make(chan *websocket.Conn)

func Run() {
	var sockets = make(map[*websocket.Conn]bool)
	var bc = context.Background()
	for {
		select {
		case cmd := <-commands:
			for s := range sockets {
				if err := wsjson.Write(bc, s, cmd); err != nil {
					log.Error("Err during write:", err)
					delete(sockets, s)
					s.CloseNow()
				}
			}
		case s := <-socketsChan:
			sockets[s] = true
		}
	}
}
