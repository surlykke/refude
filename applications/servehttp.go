package applications

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetJsonResource(r *http.Request) respond.JsonResource {
	if strings.HasPrefix(r.URL.Path, "/application/") {
		if app, ok := collectionStore.Load().(collection).applications[r.URL.Path[13:]]; ok {
			return app
		}
	} else if strings.HasPrefix(r.URL.Path, "/mimetype/") {
		if mt, ok := collectionStore.Load().(collection).mimetypes[r.URL.Path[10:]]; ok {
			return mt
		}
	}
	return nil
}

func Crawl(term string, forDisplay bool, crawler searchutils.Crawler) {
	if term == "" {
		return
	}

	if !forDisplay {
		for _, mt := range collectionStore.Load().(collection).mimetypes {
			crawler(&mt.Resource, nil)
		}
	}

	for _, app := range collectionStore.Load().(collection).applications {
		if !(forDisplay && app.NoDisplay) {
			crawler(&app.Resource, app.Keywords)
		}
	}
}
