// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package browser

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/log"
	"github.com/surlykke/refude/internal/lib/mediatype"
	"github.com/surlykke/refude/internal/lib/pubsub"
	"github.com/surlykke/refude/internal/lib/repo"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
	"github.com/surlykke/refude/internal/watch"
)

var TabMap = repo.MakeSynkMap[string, *Tab]()
var BookmarkMap = repo.MakeSynkMap[string, *Bookmark]()

type browserCommand struct {
	BrowserId string `json:"browserId"`
	Cmd       string `json:"cmd"` // "focus" or "close"
	TabId     string `json:"tabId"`
}

var browserCommands = pubsub.MakePublisher[browserCommand]()

type postData struct {
	Type string `json:"type"` // "tabs" or "bookmarks"
	List []struct {
		Id      string    `json:"id"`
		Url     string    `json:"url"`
		Title   string    `json:"title"`
		Favicon icon.Name `json:"favicon"`
	}
}

/*func BookmarksDoPost(dataList []sinkData) response.Response {
	var mapOfBookmarks = make(map[string]*Bookmark, len(dataList))
	for _, data := range dataList {
		if data.Url == "" {
			continue
		}
		var iconUrl = icon.Name("https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=" + url.QueryEscape(data.Url))
		var bookMark = Bookmark{Base: *entity.MakeBase(data.Title, " ", iconUrl, mediatype.Bookmark), Id: data.Id, ExternalUrl: data.Url}
		bookMark.AddAction("", "Bookmark", "")
		mapOfBookmarks[data.Id] = &bookMark

	}
	BookmarkMap.ReplaceAll(mapOfBookmarks)
	return response.Accepted()
}*/

func Run() {
	os.Remove(xdg.NmSocketPath)
	if listener, err := net.Listen("unix", xdg.NmSocketPath); err != nil {
		log.Warn(err)
		return
	} else {
		for {
			if conn, err := listener.Accept(); err != nil {
				log.Warn(err)
			} else {
				go receiver(conn)
			}
		}
	}
}

func receiver(conn net.Conn) {
	defer conn.Close()
	if data, err := readMsg(conn); err != nil {
		return
	} else {
		var browserId = string(data)
		go sender(browserId, conn)

		for {
			var brd postData
			if data, err := readMsg(conn); err != nil {
				log.Warn(err)
				return
			} else if err := json.Unmarshal(data, &brd); err != nil {
				log.Warn("Invalid json:\n", string(data))
			} else {
				if brd.Type == "tabs" {
					var mapOfTabs = make(map[string]*Tab, len(brd.List))
					for _, d := range brd.List {
						if len(d.Title) > 60 { // Shorten title a bit
							d.Title = d.Title[0:60] + "..."
						}
						var tab = &Tab{Base: *entity.MakeBase(d.Title, browserId+" tab", d.Favicon, mediatype.Tab), Id: d.Id, BrowserId: browserId, Url: d.Url}
						tab.AddAction("", browserId+" tab", "")
						mapOfTabs[d.Id] = tab
					}
					TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == browserId })
					watch.Publish("search", "")
				}
			}
		}
	}
}

func readMsg(conn net.Conn) ([]byte, error) {
	var sizeBuf = make([]byte, 4)
	var dataBuf = make([]byte, 65536)
	var size uint32
	if n, err := conn.Read(sizeBuf); err != nil || n < 4 {
		return nil, errors.Wrap(err, fmt.Sprintf("read: %d", n))
	} else if _, err = binary.Decode(sizeBuf, binary.NativeEndian, &size); err != nil {
		return nil, err
	} else {
		var read uint32 = 0
		for read < size {
			if n, err := conn.Read(dataBuf[read:]); err != nil {
				return nil, err
			} else {
				read = read + uint32(n)
			}
		}
		return dataBuf[0:size], nil
	}
}

func sender(browserId string, conn net.Conn) {
	var subscription = browserCommands.Subscribe()
	for {
		var cmd = subscription.Next()
		if cmd.BrowserId == browserId {
			b, _ := json.Marshal(cmd)
			if _, err := conn.Write(utils.PrependWithLength(b)); err != nil {
				return
			}
		}
	}
}
