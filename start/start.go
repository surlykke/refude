// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package start

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/file"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/windows"
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
	var sources = make([]chan link.Link, 0, 12)

	var addSources = func(searchFuncs...func(chan link.Link, string)) {
		for _, searchFunc := range searchFuncs {
			var sink = make(chan link.Link)
			go searchFunc(sink, term)  
			sources = append(sources, sink)
		}
	}
	
	addSources(notifications.Search, windows.WM.Search)

	if len(term) > 0 {
		addSources(applications.Search, file.Search)
	}

	if len(term) > 3 {
		addSources(power.Search)
	}

	for _, ch := range sources {
		for l := range ch {
			links = append(links, l)
		}
	}

	return links
}


