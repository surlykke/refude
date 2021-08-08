package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)

func Filter(links []resource.Link, term string) []resource.Link {

	var result = make([]resource.Link, 1, len(links))
	result[0] = links[0]
	result[0].Rank = 0
	for _, l := range links[1:] {
		// Links here should be newly minted, so writing them is ok.
		if l.Rank = searchutils.Match(term, l.Title); l.Rank > -1 {
			if l.Relation == relation.Related {
				l.Rank += 10
			}
			result = append(result, l)
		}
	}
	sort.Sort(resource.LinkList(result[1:]))
	return result
}

func doDesktopSearch(term string) []resource.Link {
	var sink = make(chan resource.Link)
	var done = make(chan struct{})
	var collectors []func(string, chan resource.Link)

	collectors = append(collectors, notifications.Collect, windows.Collect)
	if len(term) > 0 {
		collectors = append(collectors, applications.CollectLinks)
		if len(term) > 3 {
			collectors = append(collectors, file.Collect, power.Collect)
		}
	}

	for _, collector := range collectors {
		var c = collector
		go func() {
			c(term, sink)
			done <- struct{}{}
		}()
	}

	var links = make([]resource.Link, 1, 1000)
	links[0] = resource.MakeLink("/desktop/search?term="+term, "Desktop Search", "", relation.Self)

	for n := len(collectors); n > 0; {
		select {
		case link := <-sink:
			links = append(links, link)
		case _ = <-done:
			n--
		}
	}
	sort.Sort(resource.LinkList(links[1:]))
	return links
}

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon", "/search/desktop", "/search/paths", "/watch", "/doc")

	var sink, done = make(chan string), make(chan struct{})

	var collectors = []func(string, chan string){
		windows.CollectPaths, applications.CollectPaths, icons.CollectPaths, statusnotifications.CollectPaths, notifications.CollectPaths, power.CollectPaths,
	}

	for _, collector := range collectors {
		var c = collector
		go func() {
			c("", sink)
			done <- struct{}{}
		}()
	}

	for count := len(collectors); count > 0; {
		select {
		case path := <-sink:
			if strings.HasPrefix(path, prefix) {
				paths = append(paths, path)
			}
		case _ = <-done:
			count--
		}
	}
	return paths
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/search/desktop" {
		if r.Method == "GET" {
			var term = requests.Term(r)
			respond.ResourceAsJson(w, doDesktopSearch(term), "search", struct{ Term string }{term})
		} else {
			respond.NotAllowed(w)
		}
	} else if r.URL.Path == "/search/paths" {
		if r.Method == "GET" {
			respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}
