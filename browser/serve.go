package browser

import (
	"net/url"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/watch"
)

var TabMap = repo.MakeSynkMap[string, *Tab]()
var BookmarkMap = repo.MakeSynkMap[string, *Bookmark]()

type sinkData struct {
	Id      string    `json:"id"`
	Url     string    `json:"url"`
	Title   string    `json:"title"`
	Favicon icon.Name `json:"favicon"`
}

func TabsDoPost(browserName string, dataList []sinkData) response.Response {
	var mapOfTabs = make(map[string]*Tab, len(dataList))
	for _, d := range dataList {
		if len(d.Title) > 60 { // Shorten title a bit
			d.Title = d.Title[0:60] + "..."
		}
		var tab = &Tab{Base: *entity.MakeBase(d.Title, d.Favicon, mediatype.Tab), Id: d.Id, BrowserId: browserName, Url: d.Url}
		tab.AddAction("", browserName+" tab", "")
		//tab.AddDeleteAction("close", title, "Close tab", "")

		mapOfTabs[d.Id] = tab
	}

	TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == browserName })
	watch.Publish("search", "")
	return response.Accepted()
}

func BookmarksDoPost(dataList []sinkData) response.Response {
	var mapOfBookmarks = make(map[string]*Bookmark, len(dataList))
	for _, data := range dataList {
		if data.Url == "" {
			continue
		}
		var iconUrl = icon.Name("https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=" + url.QueryEscape(data.Url))
		var bookMark = Bookmark{Base: *entity.MakeBase(data.Title, iconUrl, mediatype.Bookmark), Id: data.Id, ExternalUrl: data.Url}
		bookMark.AddAction("", "Bookmark", "")
		mapOfBookmarks[data.Id] = &bookMark

	}
	BookmarkMap.ReplaceAll(mapOfBookmarks)
	return response.Accepted()
}
