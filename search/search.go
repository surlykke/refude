package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)

func doDesktopSearch(term string) link.List {
	var sink = make(chan link.Link)
	var done = make(chan struct{})
	var collectors []func(string, chan link.Link)

	collectors = append(collectors, notifications.Notifications.Search, windows.Windows.Search)
	if len(term) > 0 {
		collectors = append(collectors, applications.Applications.Search)
		if len(term) > 3 {
			collectors = append(collectors, file.Collect, power.Devices.Search)
		}
	}

	for _, collector := range collectors {
		var c = collector
		go func() {
			c(term, sink)
			done <- struct{}{}
		}()
	}

	var links = make(link.List, 0, 1000)
	for n := len(collectors); n > 0; {
		select {
		case link := <-sink:
			links = append(links, link)
		case _ = <-done:
			n--
		}
	}
	sort.Sort(links)
	return links
}

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon", "/search/desktop", "/search/paths", "/watch", "/doc")

	var sink, done = make(chan string), make(chan struct{})

	var collectors = []func(string, chan string){
		windows.Windows.Paths, applications.Applications.Paths, icons.CollectPaths, statusnotifications.Items.Paths, notifications.Notifications.Paths, power.Devices.Paths,
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

type search struct {
	links link.List
	Term  string
}

func (s search) Links(path string) link.List {
	return s.links
}

func (s search) ForDisplay() bool {
	return false
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/search/desktop" {
		if r.Method == "GET" {
			var term = requests.Term(r)
			resource.Make("/search/desktop", "Search", "", "", "search", search{links: doDesktopSearch(term), Term: term}).ServeHTTP(w, r)
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
