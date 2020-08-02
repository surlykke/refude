package applications

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
)

var appPattern = regexp.MustCompile("^/application/([^/]+)(/actions|/action/([^/]+))?$")
var mimePattern = regexp.MustCompile("^/mimetype/(.+)$")

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/applications" {
		return Applications()
	} else if strings.HasPrefix(r.URL.Path, "/application/") {
		if app, ok := collectionStore.Load().(collection).applications[r.URL.Path[13:]]; ok {
			return app
		}
	} else if r.URL.Path == "/mimetypes" {
		return Mimetypes()
	} else if strings.HasPrefix(r.URL.Path, "/mimetype/") {
		if mt, ok := collectionStore.Load().(collection).mimetypes[r.URL.Path[10:]]; ok {
			return mt
		}
	}

	return nil
}

func Applications() respond.StandardFormatList {
	var c = collectionStore.Load().(collection)
	var sfl = make(respond.StandardFormatList, 0, len(c.applications))
	for _, app := range c.applications {
		sfl = append(sfl, app.ToStandardFormat())
	}
	return sfl
}

func Mimetypes() respond.StandardFormatList {
	var c = collectionStore.Load().(collection)
	var sfl = make(respond.StandardFormatList, 0, len(c.mimetypes))
	for _, mt := range c.mimetypes {
		sfl = append(sfl, mt.ToStandardFormat())
	}
	return sfl
}

func AllPaths() []string {
	var c = collectionStore.Load().(collection)
	var paths = make([]string, 0, len(c.applications)+len(c.mimetypes)+100)
	for _, app := range c.applications {
		paths = append(paths, app.self)
	}
	for _, mt := range c.mimetypes {
		paths = append(paths, mt.self)
	}
	paths = append(paths, "/applications", "/mimetypes")
	return paths
}
