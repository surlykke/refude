package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/file"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/session"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func Handler(r *http.Request) http.Handler {
	if r.URL.Path == "/search/paths" {
		return PathSearcher{}
	} else if r.URL.Path == "/search/desktop" {
		return DesktopSearcher{}
	} else {
		return nil
	}
}

type PathSearcher struct{}

func (p PathSearcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var prefix = requests.GetSingleQueryParameter(r, "prefix", "")
		var paths = make([]string, 0, 2000)

		paths = append(paths, windows.AllPaths()...)
		paths = append(paths, applications.AllPaths()...)
		paths = append(paths, icons.AllPaths()...)
		paths = append(paths, statusnotifications.AllPaths()...)
		paths = append(paths, session.AllPaths()...)
		paths = append(paths, notifications.AllPaths()...)
		paths = append(paths, power.AllPaths()...)
		paths = append(paths, otherPaths()...)

		var found = 0
		for i := 0; i < len(paths); i++ {
			if strings.HasPrefix(paths[i], prefix) {
				paths[found] = paths[i]
				found++
			}
		}
		paths = paths[:found]

		sort.Sort(slice.SortableStringSlice(paths))
		respond.AsJson2(w, paths)
	} else {
		respond.NotAllowed(w)
	}
}

type DesktopSearcher struct{}

func (d DesktopSearcher) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		var term = requests.Term(r)

		var sfl = make(respond.StandardFormatList, 0, 1000)
		if term == "" {
			sfl = append(sfl, file.Recent().Filter("")...)
			sfl = append(sfl, notifications.CollectActionable().Filter("")...)
			sfl = append(sfl, windows.Windows().Filter("")...)
		} else {
			sfl = append(sfl, notifications.CollectActionable().Filter(term).Sort()...)
			sfl = append(sfl, windows.Windows().Filter(term).Sort()...)
			sfl = append(sfl, applications.Applications().Filter(term).Sort()...)
			sfl = append(sfl, session.Collect().Filter(term).Sort()...)
			sfl = append(sfl, file.DesktopSearch(term).Filter(term).Sort()...)
			sfl = append(sfl, power.DesktopSearch().Filter(term).Sort()...)
		}

		respond.AsJson2(w, sfl)
	} else {
		respond.NotAllowed(w)
	}
}

func otherPaths() []string {
	return []string{"/search/paths", "/search/desktop", "/watch"}
}
