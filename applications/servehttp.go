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

func CollectLinks(term string, sink chan resource.Link) {
	for _, app := range collectionStore.Load().(collection).applications {
		if !app.NoDisplay {
			if rnk := searchutils.Match(term, app.Name, app.Keywords...); rnk > -1 {
				sink <- resource.MakeRankedLink(app.self, app.Name, app.Icon, "application", rnk)
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
