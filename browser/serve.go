package browser

import (
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
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

func TabsDoPost(browserId string, dataList []sinkData) response.Response {
	if browserTitle, _, ok := applications.GetTitleAndIcon(browserId); ok {
		var mapOfTabs = make(map[string]*Tab, len(dataList))
		for _, d := range dataList {
			if len(d.Title) > 60 { // Shorten title a bit
				d.Title = d.Title[0:60] + "..."
			}
			var tab = &Tab{Base: *entity.MakeBase(d.Title, browserTitle+" tab", d.Favicon, mediatype.Tab), Id: d.Id, BrowserId: browserId, Url: d.Url}
			tab.AddAction("", browserTitle+" tab", "")
			mapOfTabs[d.Id] = tab
		}

		TabMap.Replace(mapOfTabs, func(t *Tab) bool { return t.BrowserId == browserId })
		watch.Publish("search", "")
		return response.Accepted()
	} else {
		return response.NotFound()
	}
}

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

/*
*

		Try to figure which browser from the useragent.
	    Alas google-chrome and chromium have identical useragents, so if both chrome and chromium are installed,
	    all tabs will appear as chrome tabs, and activating a tab will open/focus it in chrome.

		Examples
		Chrome: 	Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36
		Chromium:	Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36
		Brave: 		Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/138.0.0.0 Safari/537.36
		Firefox: 	Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:140.0) Gecko/20100101 Firefox/140.0
		Edge: 		Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 Edg/129.0.0.0

		Rules (taken from MDN)

		Browser name	Must contain			Must not contain
		----------------------------------------------------------------
		Firefox			Firefox/xyz				Seamonkey/xyz
		Seamonkey		Seamonkey/xyz
		Chrome			Chrome/xyz				Chromium/xyz or Edg./xyz // NOTE No longer correct
		Chromium		Chromium/xyz                                     //  -----  "  -----------
		Safari			Safari/xyz *			Chrome/xyz or Chromium/xyz
		Opera 15+		(Blink-based engine)	OPR/xyz
		Opera 12-		(Presto-based engine)	Opera/xyz
*/
func GuessBrowser(userAgent string) (appId, name string) {
	if strings.Contains(userAgent, "Firefox/") {
		return firefoxId, firefoxName
	} else if strings.Contains(userAgent, "Chrome/") {
		if strings.Contains(userAgent, "Edg/") {
			return edgeId, edgeName
		} else {
			return chromeId, chromeName
		}
	} else {
		return "", ""
	}
}

var chromeId, chromeName = "google-chrome", "Google Chrome" // TODO
var firefoxId, firefoxName = "firefox_firefox", "Firefox"
var edgeId, edgeName = "microsoft-edge", "Microsoft Edge"

/**
 */

/**
 */
