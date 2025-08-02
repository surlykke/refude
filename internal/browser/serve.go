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
	"io"
	"log"
	"net"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/pubsub"
	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
	"github.com/surlykke/refude/internal/watch"
)

var TabMap = entity.MakeMap[string, *Tab]()
var BookmarkMap = entity.MakeMap[string, *Bookmark]()

// Data sent to the browser
type browserCommand struct {
	BrowserId string `json:"browserId"`
	Cmd       string `json:"cmd"` // "report", "focus" or "close"
	TabId     string `json:"tabId"`
}

var browserCommands = pubsub.MakePublisher[browserCommand]()

// Data comming from the browser
type browserData struct {
	Type string `json:"type"` // "tabs" or "bookmarks"
	List []struct {
		Id      string `json:"id"`
		Url     string `json:"url"`
		Title   string `json:"title"`
		Favicon string `json:"favicon"`
	}
}

/*func BookmarksDoPost(dataList []sinkData) response.Response {
	var mapOfBookmarks = make(map[string]*Bookmark, len(dataList))
	for _, data := range dataList {
		if data.Url == "" {
			continue
		}
		var iconUrl = icon.Name("https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=" + url.QueryEscape(data.Url))
		var bookMark = Bookmark{Base: *entity.MakeBase(data.Title, " ", iconUrl, entity.Bookmark), Id: data.Id, ExternalUrl: data.Url}
		bookMark.AddAction("", "Bookmark", "")
		mapOfBookmarks[data.Id] = &bookMark

	}
	BookmarkMap.ReplaceAll(mapOfBookmarks)
	return response.Accepted()
}*/

func Run() {
	os.Remove(xdg.NmSocketPath)
	if listener, err := net.Listen("unix", xdg.NmSocketPath); err != nil {
		log.Print(err)
		return
	} else {
		for {
			if conn, err := listener.Accept(); err != nil {
				log.Print(err)
			} else {
				go receive(conn)
			}
		}
	}
}

func receive(conn net.Conn) {
	defer conn.Close()
	if data, err := readMsg(conn); err != nil {
		return
	} else {
		var browserId = string(data)
		defer clean(browserId)
		var browserName = browserNameFromId(browserId)
		log.Print("Connected to ", browserName)
		go send(browserId, conn)

		for {
			var brd browserData
			if data, err := readMsg(conn); err == io.EOF {
				log.Print("Disconnected from ", browserName)
				return
			} else if err != nil {
				log.Print(err, "- disconnecting from", browserName)
				return
			} else if err := json.Unmarshal(data, &brd); err != nil {
				log.Print("Invalid json:\n", string(data))
			} else if brd.Type == "tabs" {
				var mapOfTabs = make(map[string]*Tab, len(brd.List))
				for _, d := range brd.List {
					if len(d.Title) > 60 { // Shorten title a bit
						d.Title = d.Title[0:60] + "..."
					}
					var tab = &Tab{Base: *entity.MakeBase(d.Title, browserName+" tab", d.Favicon, "Browser tab"), Id: d.Id, BrowserId: browserId, Url: d.Url}
					tab.AddAction("", browserId+" tab", "")
					mapOfTabs[d.Id] = tab
				}
				TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == browserId })
				watch.Publish("search", "")
			} // TODO: bookmarks
		}
	}
}

func readMsg(conn net.Conn) ([]byte, error) {
	var sizeBuf = make([]byte, 4)
	var dataBuf = make([]byte, 65536)
	var size uint32
	if n, err := conn.Read(sizeBuf); err != nil {
		if err != io.EOF {
			log.Print("readMsg, err:", err)
		}
		return nil, err
	} else if n < 4 {
		return nil, errors.New(fmt.Sprintf("Expected at least 4 bytes, got: %d", n))
	} else {
		binary.Decode(sizeBuf, binary.NativeEndian, &size) // Only errors if buf too small
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

func browserNameFromId(id string) string {
	if strings.Contains(id, "chrome") {
		return "Chrome"
	} else if strings.Contains(id, "msedge") {
		return "Edge"
	} else if strings.Contains(id, "brave") {
		return "Brave"
	} else if strings.Contains(id, "vivaldi") {
		return "Vivaldi"
	} else {
		// TODO: chromium, firefox...
		return id
	}

}

var reportCommand = response.ToJson(browserCommand{Cmd: "report"})

func send(browserId string, conn net.Conn) {
	var subscription = browserCommands.Subscribe()
	if err := writeMsg(conn, reportCommand); err != nil {
		log.Print(err)
		return
	}
	for {
		var cmd = subscription.Next()
		if cmd.BrowserId == browserId {
			if err := writeMsg(conn, response.ToJson(cmd)); err != nil {
				return
			}
		}
	}
}

func writeMsg(conn net.Conn, msg []byte) error {
	var _, err = conn.Write(utils.PrependWithLength(msg))
	if err != nil && err != io.ErrClosedPipe {
		log.Print(err)
	}
	return err
}

func clean(browserId string) {
	TabMap.Replace(map[string]*Tab{}, func(t *Tab) bool { return t.BrowserId == browserId })
}
