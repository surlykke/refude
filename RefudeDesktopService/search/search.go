package search

import (
	"net/http"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/notifications"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/power"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/session"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/windows"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/statusnotifications"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/backlight"
	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/applications"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/complete":
		fallthrough // deprecated
	case "/search/paths":
		searchPaths(w, r)
	case "/search/desktop":
		searchDesktop(w, r)
	default:
		respond.NotFound(w)
	}
}

func searchPaths(w http.ResponseWriter, r *http.Request) {

	var prefix = requests.GetSingleQueryParameter(r, "prefix", "")
	var paths = make([]string, 0, 2000)

	paths = append(paths, windows.AllPaths()...)
	paths = append(paths, applications.AllPaths()...)
	paths = append(paths, backlight.AllPaths()...)
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
	respond.AsJson(w, r, paths)
}

func searchDesktop(w http.ResponseWriter, r *http.Request) {
	var term = searchutils.Term(r)

	var sfl = make(respond.StandardFormatList, 0, 1000)
	if term == "" {
		sfl = append(sfl, notifications.Collect("")...)
		sfl = append(sfl, windows.Collect("")...)
	} else {
		sfl = append(sfl, notifications.Collect("")...)
		sfl = append(sfl, windows.Collect(term)...)
		sfl = append(sfl, applications.CollectApps(term)...)
		sfl = append(sfl, session.Collect(term)...)
	}

	var j = 0
	for i := 0; i < len(sfl); i++ {
		if !sfl[i].NoDisplay {
			sfl[j] = sfl[i]
			j++
		}
	}

	sfl = sfl[:j]

	respond.AsJson(w, r, sfl)
}

func otherPaths() []string {
	return []string{"/search/paths", "/search", "/search/events", "/search/desktop", "/complete"}
}
