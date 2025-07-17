package browser

import (
	"net/url"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/response"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func BookmarksDoPost(dataList []sinkData) response.Response {
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
}

type Bookmark struct {
	entity.Base
	Id          string
	ExternalUrl string
}

func (this *Bookmark) DoPost(action string) response.Response {
	xdg.RunCmd("xdg-open", this.ExternalUrl)
	return response.Accepted()
}

// We use this for icon url
// https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=<bookmark url>
