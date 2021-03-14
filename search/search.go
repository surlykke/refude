package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"

	"github.com/surlykke/RefudeServices/statusnotifications"

	"github.com/surlykke/RefudeServices/icons"

	"github.com/surlykke/RefudeServices/applications"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type SearchResult struct {
	respond.Resource
	Term string `json:"term"`
}

func makeSearchResult(term string, links []respond.Link) *SearchResult {
	var sr = SearchResult{
		Term: term,
	}
	sr.Resource = respond.MakeResource("/desktop/search?term="+term, "Desktop Search", "", &sr, "search")
	sr.AddLink(links...)
	return &sr
}

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/search/paths" {
		return http.HandlerFunc(Paths)
	} else if r.URL.Path == "/search/desktop" {
		return http.HandlerFunc(DesktopResources)
	} else {
		return nil
	}
}

func Paths(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var prefix = requests.GetSingleQueryParameter(r, "prefix", "")
		var paths = make([]string, 0, 2000)

		paths = append(paths, windows.AllPaths()...)
		paths = append(paths, applications.AllPaths()...)
		paths = append(paths, icons.AllPaths()...)
		paths = append(paths, statusnotifications.AllPaths()...)
		paths = append(paths, notifications.AllPaths()...)
		paths = append(paths, power.AllPaths()...)
		paths = append(paths, "/search/paths", "/search/desktop", "/watch")

		var found = 0
		for i := 0; i < len(paths); i++ {
			if strings.HasPrefix(paths[i], prefix) {
				paths[found] = paths[i]
				found++
			}
		}
		paths = paths[:found]

		sort.Sort(slice.SortableStringSlice(paths))
		respond.AsJson(w, paths)
	} else {
		respond.NotAllowed(w)
	}
}

func DesktopResources(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var term = strings.ToLower(requests.Term(r))

		var sfl = make([]respond.Link, 0, 1000)
		sfl = append(sfl, file.Recent(term, 0)...)
		sfl = append(sfl, notifications.DesktopSearch(term, 100)...)
		sfl = append(sfl, windows.DesktopSearch(term, 200)...)

		var pos = len(sfl)

		if len(term) > 0 {
			sfl = append(sfl, applications.DesktopSearch(term, 300)...)
			sfl = append(sfl, file.DesktopSearch(term, 300)...)
			sfl = append(sfl, power.DesktopSearch(term, 600)...)
		}
		if len(sfl) > pos { // We do not want to change order of files, notifikations and windows, so no sorting for them
			sort.Sort(respond.LinkList(sfl[pos:]))
		}
		respond.AsJson(w, makeSearchResult(term, sfl))
	} else {
		respond.NotAllowed(w)
	}
}
