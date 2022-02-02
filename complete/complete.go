// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package complete

import (
	"net/http"
	"strings"

	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)


func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	for _, path := range []string{"/icon?name=", "/start?want=links&term=", "/complete?prefix=", "/watch", "/doc"} {
		if strings.HasPrefix(path, prefix) {
			paths = append(paths, path)
		}
	}

	for _, list := range []*resource.List{windows.Windows, applications.Applications, applications.Mimetypes, statusnotifications.Items, notifications.Notifications, power.Devices, icons.IconThemes} {
		for _, res := range list.GetAll() {
			if strings.HasPrefix(string(res.Self), prefix) {
				paths = append(paths, string(res.Self))
			}
		}
	}

	return paths
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
	} else {
		respond.NotAllowed(w)
	}
}
