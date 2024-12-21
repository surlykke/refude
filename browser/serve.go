package browser

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/watch"
)

// This is the api that the browserextensions use
func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		if r.Body == nil {
			respond.UnprocessableEntity(w, errors.New("No data"))
		} else if bytes, err := io.ReadAll(r.Body); err != nil {
			respond.ServerError(w, err)
		} else {
			var data = make([]map[string]string, 30)
			if err := json.Unmarshal(bytes, &data); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if r.URL.Path == "/tabsink" {
				var tabs = make([]resource.Resource, 0, len(data))
				for _, d := range data {
					var title = d["title"]
					if len(title) > 60 { // Shorten title a bit
						if spacePos := strings.Index(title[60:], " "); spacePos > -1 {
							title = title[0:60+spacePos] + "..."
						} else {
							title = title[0:60] + "..."
						}
					}
					var url = d["url"]
					if queryStart := strings.Index(url, "?"); queryStart > -1 {
						url = url[0:queryStart+1] + "..."
					} else if len(url) > 60 {
						url = url[0:60] + "..."
					}
					var iconUrl = d["favIcon"]
					var tab = &Tab{ResourceData: *resource.MakeBase(path.Of("/tab/", d["id"]), title, "", icon.Name(iconUrl), mediatype.Tab)}
					tab.Url = url
					tab.AddAction("focus", title, "Focus tab", icon.Name(iconUrl))
					//tab.AddDeleteAction("close", title, "Close tab", "")

					tabs = append(tabs, tab)
				}
				respond.Ok(w)
				repo.Replace(tabs, "/tab/")
				watch.Publish("search", "")
			} else { // /bookmarksink
				var bookmarks = make([]resource.Resource, 0, len(data))
				for _, d := range data {
					var externalUrl = d["url"]
					if externalUrl == "" {
						continue
					}
					var iconUrl = "https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=" + url.QueryEscape(externalUrl)

					var baseData = resource.MakeBase(path.Of("/bookmark/", d["id"]), d["title"], "", icon.Name(iconUrl), mediatype.Bookmark)
					var bookMark = Bookmark{ResourceData: *baseData, Id: d["id"], ExternalUrl: externalUrl}
					bookMark.AddAction("open", d["title"], "Open bookmark", icon.Name(iconUrl))
					bookmarks = append(bookmarks, &bookMark)
				}
				repo.Replace(bookmarks, "/bookmark/")
				watch.Publish("search", "")
			}
		}
	} else {
		respond.NotAllowed(w)
	}
}
