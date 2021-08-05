package applications

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetAppResource(pathComponents []string) resource.Resource {
	if len(pathComponents) == 1 {
		if app, ok := collectionStore.Load().(collection).applications[pathComponents[0]]; ok {
			return app
		}
	}
	return nil
}

func GetMimeResource(relPath []string) resource.Resource {
	if len(relPath) == 1 {
		if mt, ok := collectionStore.Load().(collection).mimetypes[relPath[0]]; ok {
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
			crawler(mt.self, mt.Comment, "")
		}
	}

	for _, app := range collectionStore.Load().(collection).applications {
		if !(forDisplay && app.NoDisplay) {
			crawler(app.self, app.Name, app.Icon, app.Keywords...)
		}
	}
}
