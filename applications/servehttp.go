package applications

import (
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

func GetAppResource(appId string) resource.Resource {
	if app, ok := collectionStore.Load().(collection).applications[appId]; ok {
		return app
	}
	return nil
}

func GetMimeResource(mimeId string) resource.Resource {
	if mt, ok := collectionStore.Load().(collection).mimetypes[mimeId]; ok {
		return mt
	}
	return nil
}

func CollectLinks(term string, sink chan link.Link) {
	for _, app := range collectionStore.Load().(collection).applications {
		if !app.NoDisplay {
			if rnk := searchutils.Match(term, app.Name, app.Keywords...); rnk > -1 {
				sink <- link.MakeRanked(app.self, app.Name, app.Icon, "application", rnk)
			}
		}
	}
}

func CollectPaths(method string, sink chan string) {
	for _, app := range collectionStore.Load().(collection).applications {
		sink <- app.self
	}
	for _, mt := range collectionStore.Load().(collection).mimetypes {
		sink <- mt.self
	}
}
