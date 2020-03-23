package search

import (
	"fmt"
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

var pathGetters = map[string]func() []string{
	"window":           windows.AllPaths,
	"application":      applications.AllAppPaths,
	"mimetype":         applications.AllMimetypePaths,
	"backlight_device": backlight.AllPaths,
	"icontheme":        icons.AllPaths,
	"status_item":      statusnotifications.AllPaths,
	"session_action":   session.AllPaths,
	"notification":     notifications.AllPaths,
	"device":           power.AllPaths,
	"other":            otherPaths,
}

var searchers = map[string]func(*searchutils.Collector){
	"window":           windows.SearchWindows,
	"application":      applications.SearchApps,
	"mimetype":         applications.SearchMimetypes,
	"backlight_device": backlight.SearchBacklights,
	"icontheme":        icons.SearchThemes,
	"status_item":      statusnotifications.SearchItems,
	"session_action":   session.SearchActions,
	"notification":     notifications.SearchNotifications,
	"device":           power.SearchDevices,
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/complete": // To be removed
		fallthrough
	case "/search/paths":
		searchPaths(w, r)
	case "/search":
		search(w, r)
	case "/search/events":
		searchEvents(w, r)
	case "/search/desktop":
		searchDesktop(w, r)
	default:
		respond.NotFound(w)
	}
}

func searchPaths(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else {
		var prefix = requests.GetSingleQueryParameter(r, "prefix", "")
		var paths = make([]string, 0, 2000)

		for _, pathGetter := range pathGetters {
			for _, path := range pathGetter() {
				if strings.HasPrefix(path, prefix) {
					paths = append(paths, path)
				}
			}
		}

		sort.Sort(slice.SortableStringSlice(paths))
		respond.AsJson(w, paths)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
		return
	}

	var query = r.URL.Query()
	var resourceTypes = query["type"]
	if len(resourceTypes) == 0 {
		respond.AsJson(w, []string{})
	}

	for _, resourceType := range resourceTypes {
		if _, ok := searchers[resourceType]; !ok {
			respond.UnprocessableEntity(w, fmt.Errorf("Unknown resource type: %s", resourceType))
			return
		}
	}
	var searchTerms = query["term"]
	if len(searchTerms) > 1 {
		respond.UnprocessableEntity(w, fmt.Errorf("Only one searchterm allowed"))
		return
	}
	var searchTerm = ""
	if len(searchTerms) > 0 {
		searchTerm = searchTerms[0]
	}

	respond.AsJson(w, find(resourceTypes, searchTerm))
}

func searchEvents(w http.ResponseWriter, r *http.Request) {
	var collector = searchutils.MakeCollector("", 20, false)
	notifications.SearchNotifications(collector)
	respond.AsJson(w, collector.Get())
}

func searchDesktop(w http.ResponseWriter, r *http.Request) {
	var terms = r.URL.Query()["term"]
	var term = ""
	if len(terms) > 1 {
		respond.UnprocessableEntity(w, fmt.Errorf("More than one searchterm"))
		return
	} else if len(terms) == 1 {
		term = terms[0]
	}
	var resourceTypes []string
	if term == "" {
		resourceTypes = []string{"notification", "window"}
	} else {
		resourceTypes = []string{"notification", "window", "application", "session_action"}
	}

	var collector = searchutils.MakeCollector(term, 200, true)
	var resources = make([]*respond.StandardFormat, 0, 1000)
	for _, resourceType := range resourceTypes {
		collector.Clear()
		searchers[resourceType](collector)
		if resourceType == "notification" || resourceType == "window" {
			resources = append(resources, collector.Get()...)
		} else {
			resources = append(resources, collector.SortByRankAndGet()...)
		}
	}

	respond.AsJson(w, resources)
}

// Caller ensures validity of resourceTypes
func find(resourceTypes []string, term string) []*respond.StandardFormat {
	var resources = make([]*respond.StandardFormat, 0, 1000)
	var collector = searchutils.MakeCollector(strings.ToLower(term), 1000, false)

	for _, resourceType := range resourceTypes {
		collector.Clear()
		searchers[resourceType](collector)
		if resourceType == "window" {
			resources = append(resources, collector.Get()...)
		} else {
			resources = append(resources, collector.SortByRankAndGet()...)
		}
	}

	return resources

}

func otherPaths() []string {
	return []string{"/search/paths", "/search", "/search/events", "/search/desktop", "/complete"}
}
