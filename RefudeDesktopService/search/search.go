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
		paths = append(paths, session.AllPaths()...)
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
		var term = requests.Term(r)

		var sfl = make(respond.StandardFormatList, 0, 1000)
		sfl = append(sfl, file.Recent().Filter("")...)
		sfl = append(sfl, notifications.CollectActionable().Filter(term).ShiftRank(50)...)
		sfl = append(sfl, windows.Windows().Filter(term).ShiftRank(100)...)

		if len(term) > 0 {
			sfl = append(sfl, applications.Applications().Filter(term).ShiftRank(150)...)
			sfl = append(sfl, session.Collect().Filter(term).ShiftRank(200)...)
			sfl = append(sfl, file.DesktopSearch(term).Filter(term).ShiftRank(250)...)
			sfl = append(sfl, power.DesktopSearch().Filter(term).ShiftRank(300)...)
		}

		respond.AsJson(w, sfl.Sort())
	} else {
		respond.NotAllowed(w)
	}
}
