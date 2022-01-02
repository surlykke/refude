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
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
	"github.com/surlykke/RefudeServices/windows/x11"
)

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	links = append(links, searchCollection(notifications.Notifications.GetAll(), term, notificationFilter)...)
	links = append(links, searchCollection(windows.Windows.GetAll(), term, windowFilter)...)

	if len(term) > 0 {
		links = append(links, searchCollection(applications.Applications.GetAll(), term, applicationFilter)...)
	}

	if len(term) > 3 {
		links = append(links, file.Collect(term)...)
		links = append(links, searchCollection(power.Devices.GetAll(), term, nil)...)

	}
	sort.Sort(links)
	return links
}

func searchCollection(list []*resource.Resource, term string, filter func(*resource.Resource) bool) link.List {
	var result = make(link.List, 0, 300)
	for _, res := range list {
		if filter != nil && !filter(res) {
			continue
		} else if rnk := searchutils.Match(term, res.Title /*TODO keywords*/); rnk > -1 {
			result = append(result, res.MakeRankedLink(rnk))
		}
	}
	return result
}

func notificationFilter(r *resource.Resource) bool {
	var n = r.Data.(*notifications.Notification)
	return n.Urgency == notifications.Critical || len(n.Actions) > 0
}

func windowFilter(r *resource.Resource) bool {
	var win = r.Data.(*windows.Window)
	return win.Name != "org.refude.browser" && win.Name != "org.refude.panel" && win.State&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0
}

func applicationFilter(r *resource.Resource) bool {
	var app = r.Data.(*applications.DesktopApplication)
	return !app.NoDisplay
}

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon", "/search/desktop", "/search/paths", "/watch", "/doc")

	for _, list := range []*resource.List{windows.Windows, applications.Applications, applications.Mimetypes, statusnotifications.Items, notifications.Notifications, power.Devices, icons.IconThemes} {
		for _, res := range list.GetAll() {
			if strings.HasPrefix(string(res.Links[0].Href), prefix) {
				paths = append(paths, string(res.Links[0].Href))
			}
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

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/search/desktop" {
		if r.Method == "GET" {
			var term = requests.Term(r)
			var res = resource.MakeResource("/search/desktop", "Search", "", "", "search", search{links: doDesktopSearch(term), Term: term})
			res.ServeHTTP(w, r)
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
