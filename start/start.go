// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package start

import (
	"net/http"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"
	"github.com/surlykke/RefudeServices/windows/x11"
	"golang.org/x/exp/constraints"
	"golang.org/x/exp/slices"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	resource.ServeResource[string](w, r, "/start", Start{})
}

type Start struct{}

func (s Start) Id() string {
	return ""
}

func (s Start) Presentation() (title string, comment string, icon link.Href, profile string) {
	return "Start", "", "", "start"
}

func (s Start) Links(self, term string) link.List {
	return doDesktopSearch(term)
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	links = append(links, searchCollection(notifications.Notifications, term, notificationFilter)...)
	links = append(links, windows.Search(term)...)

	if len(term) > 0 {
		links = append(links, file.SearchFrom(xdg.Home, term, "~/")...)
		links = append(links, searchCollection(applications.Applications, term, applicationFilter)...)
	}

	if len(term) > 3 {
		links = append(links, searchCollection(power.Devices, term, nil)...)

	}

	slices.SortFunc(links, func(l1, l2 link.Link) bool {
		if l1.Rank == l2.Rank {
			return l1.Href < l2.Href
		} else {
			return l1.Rank < l2.Rank
		}
	})

	return links
}

func searchCollection[ID constraints.Ordered, T resource.Resource[ID]](collection *resource.Collection[ID, T], term string, filter func(T) bool) link.List {
	var result = make(link.List, 0, 300)
	for _, res := range collection.GetAll() {
		if filter != nil && !filter(res) {
			continue
		} else {
			var title, _, _, _ = res.Presentation()
			if rnk := searchutils.Match(term, title /*TODO keywords*/); rnk > -1 {
				result = append(result, resource.LinkTo[ID](res, collection.Prefix, rnk))
			}
		}
	}
	return result
}

func notificationFilter(n *notifications.Notification) bool {
	return n.Urgency == notifications.Critical || (len(n.NActions) > 0 && n.Urgency == notifications.Normal && n.Created+60000 > time.Now().UnixMilli())
}

func windowFilter(win windows.Window) bool {
	return win.Name != "org.refude.browser" && win.Name != "org.refude.panel" && win.State&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0
}

func applicationFilter(app *applications.DesktopApplication) bool {
	return !app.NoDisplay
}
