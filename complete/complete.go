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
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/notifications"
	"github.com/surlykke/RefudeServices/power"
	"github.com/surlykke/RefudeServices/statusnotifications"
	"github.com/surlykke/RefudeServices/windows"
)

func collectPaths(prefix string) []string {
	var paths = make([]string, 0, 1000)
	paths = append(paths, "/icon?name=", "/start?search=", "/complete?prefix=", "/watch", "/doc")
	paths = append(paths, windows.WM.GetPaths()...)
	paths = append(paths, applications.Applications.GetPaths()...)
	paths = append(paths, applications.Mimetypes.GetPaths()...)
	paths = append(paths, statusnotifications.Items.GetPaths()...)
	paths = append(paths, notifications.Notifications.GetPaths()...)
	paths = append(paths, power.Devices.GetPaths()...)
	paths = append(paths, icons.IconThemes.GetPaths()...)

	var pos = 0
	for _, path := range paths {
		if strings.HasPrefix(path, prefix) {
			paths[pos] = path
			pos = pos + 1
		}
	}

	return paths[0:pos]
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, collectPaths(requests.GetSingleQueryParameter(r, "prefix", "")))
	} else {
		respond.NotAllowed(w)
	}
}
