package applications

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
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

func Applications() respond.Links {
	var c = collectionStore.Load().(collection)
	var links = make(respond.Links, 0, len(c.applications))
	for _, app := range c.applications {
		links = append(links, app.Link())
	}
	sort.Sort(links)
	return links
}

func DesktopSearch(term string, baserank int) respond.Links {
	var applications = collectionStore.Load().(collection).applications
	var links = make(respond.Links, 0, len(applications))
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
			var link = app.Link()
			link.Rank = rank
			links = append(links, link)
		}
	}
	return links
}

func Mimetypes() respond.Links {
	var c = collectionStore.Load().(collection)
	var links = make(respond.Links, 0, len(c.mimetypes))
	for _, mt := range c.mimetypes {
		links = append(links, mt.Link())
	}
	sort.Sort(links)
	return links
}

func AllPaths() []string {
	var c = collectionStore.Load().(collection)
	var paths = make([]string, 0, len(c.applications)+len(c.mimetypes)+100)
	for _, app := range c.applications {
		paths = append(paths, app.Link().Href)
	}
	for _, mt := range c.mimetypes {
		if len(mt.Links) == 0 {
			fmt.Println("No links for", mt.Comment, mt.Id)
		}
		paths = append(paths, mt.Link().Href)
	}
	paths = append(paths, "/applications", "/mimetypes")
	return paths
}
