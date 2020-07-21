package applications

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/surlykke/RefudeServices/lib/respond"
)

var appPattern = regexp.MustCompile("^/application/([^/]+)(/actions|/action/([^/]+))?$")
var mimePattern = regexp.MustCompile("^/mimetype/(.+)$")

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/applications" {
		return Applications()
	} else if matches := appPattern.FindStringSubmatch(r.URL.Path); matches != nil {
		if app, ok := collectionStore.Load().(collection).applications[matches[1]]; !ok {
			return nil
		} else if matches[2] == "/actions" {
			return app.collectActions("")
		} else if matches[3] != "" {
			return app.DesktopActions[matches[3]]
		} else {
			return app
		}
	} else if r.URL.Path == "/mimetypes" {
		return Mimetypes()
	} else if matches := mimePattern.FindStringSubmatch(r.URL.Path); matches != nil {
		fmt.Println("Serving mimetype", matches[1])
		return collectionStore.Load().(collection).mimetypes[matches[1]]
	} else {
		return nil
	}
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
	for appId, app := range c.applications {
		paths = append(paths, appSelf(appId))
		if len(app.DesktopActions) > 0 {
			paths = append(paths, "/application/"+appId+"/actions")
			for _, act := range app.DesktopActions {
				paths = append(paths, act.self)
			}
		}
	}
	for mimetypeId := range c.mimetypes {
		paths = append(paths, mimetypeSelf(mimetypeId))
	}
	paths = append(paths, "/applications", "/mimetypes")
	return paths
}
