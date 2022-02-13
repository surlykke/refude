// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package start

import (
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/searchutils"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"
	"github.com/surlykke/RefudeServices/windows/x11"
)

type Start struct{}

func (s Start) Self() string {
	return "/start"
}

func (s Start) Presentation() (title string, comment string, icon link.Href, profile string) {
	return "Start", "", "", "start"
}

func (s Start) Links(term string) (links link.List, filtered bool) {
	return doDesktopSearch(term), true
}

func doDesktopSearch(term string) link.List {
	var links = make(link.List, 0, 300)
	term = strings.ToLower(term)

	links = append(links, searchCollection(notifications.Notifications.GetAll(), term, notificationFilter)...)
	links = append(links, searchCollection(windows.Windows.GetAll(), term, windowFilter)...)

	if len(term) > 0 {
		links = append(links, file.Search(term)...)
		links = append(links, searchCollection(applications.Applications.GetAll(), term, applicationFilter)...)
	}

	if len(term) > 3 {
		links = append(links, searchCollection(power.Devices.GetAll(), term, nil)...)

	}
	sort.Sort(links)
	return links
}

func searchCollection(list []resource.Resource, term string, filter func(resource.Resource) bool) link.List {
	var result = make(link.List, 0, 300)
	for _, res := range list {
		if filter != nil && !filter(res) {
			continue
		} else {
			var title,_,_,_ = res.Presentation()
			if rnk := searchutils.Match(term, title /*TODO keywords*/); rnk > -1 {
				result = append(result, resource.LinkTo(res, rnk))
			}
		}
	}
	return result
}

func notificationFilter(r resource.Resource) bool {
	var n = r.(*notifications.Notification)
	return n.Urgency == notifications.Critical || len(n.NActions) > 0
}

func windowFilter(r resource.Resource) bool {
	var win = r.(*windows.Window)
	return win.Name != "org.refude.browser" && win.Name != "org.refude.panel" && win.State&(x11.SKIP_TASKBAR|x11.SKIP_PAGER|x11.ABOVE) == 0
}

func applicationFilter(r resource.Resource) bool {
	var app = r.(*applications.DesktopApplication)
	return !app.NoDisplay
}

