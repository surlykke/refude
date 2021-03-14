package applications

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
)

var appPattern = regexp.MustCompile("^/application/([^/]+)(/actions|/action/([^/]+))?$")
var mimePattern = regexp.MustCompile("^/mimetype/(.+)$")

func Handler(r *http.Request) http.Handler {
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

func DesktopSearch(term string, baserank int) []respond.Link {
	var applications = collectionStore.Load().(collection).applications
	var links = make([]respond.Link, 0, len(applications))
	var termRunes = []rune(term)
	for _, app := range applications {
		if app.NoDisplay {
			continue
		}
		var rank int
		var ok bool
		var name = strings.ToLower(app.Name)
		if rank, ok = searchutils.Rank(strings.ToLower(name), term, baserank); !ok {
			if rank, ok = searchutils.Rank(strings.ToLower(app.Comment), term, baserank+100); !ok {
				rank, ok = searchutils.FluffyRank([]rune(name), termRunes, baserank+200)
			}
		}
		if ok {
			links = append(links, app.GetRelatedLink(rank))
		}
	}
	return links
}

func AllPaths() []string {
	var c = collectionStore.Load().(collection)
	var paths = make([]string, 0, len(c.applications)+len(c.mimetypes)+100)
	for _, app := range c.applications {
		paths = append(paths, app.Self.Href)
	}
	for _, mt := range c.mimetypes {
		paths = append(paths, mt.Self.Href)
	}
	paths = append(paths)
	return paths
}
