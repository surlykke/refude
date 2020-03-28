// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/surlykke/RefudeServices/lib/requests"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func AppServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/application/actionsearch" {
		if r.Method != "GET" {
			respond.NotAllowed(w)
		} else if id := requests.GetSingleQueryParameter(r, "for", ""); id == "" {
			respond.NotFound(w)
		} else {
			var prefix = "/application/" + id
			var term = requests.GetSingleQueryParameter(r, "term", "")
			var collector = searchutils.MakeCollector(term, 10, false)
			var foundSome = false
			for path, act := range actions.Load().(actionMap) {
				if strings.HasPrefix(path, prefix) {
					foundSome = true
					collector.Collect(act.ToStandardFormat())
				}
			}
			if !foundSome {
				respond.NotFound(w)
			} else {
				respond.AsJson(w, collector.SortByRankAndGet())
			}
		}
	} else if app, ok := applications.Load().(ApplicationMap)[r.URL.Path]; ok {
		if r.Method == "GET" {
			respond.AsJson(w, app.ToStandardFormat())
		} else if r.Method == "POST" {
			launch(app.Exec, app.Terminal)
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	} else if act, ok := actions.Load().(actionMap)[r.URL.Path]; ok {
		if r.Method == "GET" {
			respond.AsJson(w, act.ToStandardFormat())
		} else if r.Method == "POST" {
			launch(act.Exec, false)
			respond.Accepted(w)
		} else {
			respond.NotAllowed(w)
		}
	} else {
		respond.NotFound(w)
	}
}

func MimetypeServeHTTP(w http.ResponseWriter, r *http.Request) {
	if mt, ok := mimetypes.Load().(MimetypeMap)[r.URL.Path]; ok {
		if r.Method == "GET" {
			respond.AsJson(w, mt.ToStandardFormat())
		} else {
			respond.NotAllowed(w) // FIXME allow PATCH to set default app
		}
	} else {
		respond.NotFound(w)
	}
}

/**
 * term must be lowercase
 */
func SearchApps(collector *searchutils.Collector) {
	for _, app := range applications.Load().(ApplicationMap) {
		collector.Collect(app.ToStandardFormat())
	}
}

func SearchMimetypes(collector *searchutils.Collector) {
	for _, mt := range mimetypes.Load().(MimetypeMap) {
		collector.Collect(mt.ToStandardFormat())
	}
}

func AllAppPaths() []string {
	var applicationMap = applications.Load().(ApplicationMap)
	var paths = make([]string, 0, len(applicationMap))
	for path, _ := range applicationMap {
		paths = append(paths, path)
	}
	return paths
}

func AllMimetypePaths() []string {
	var mimetypeMap = mimetypes.Load().(MimetypeMap)
	var paths = make([]string, 0, len(mimetypeMap))
	for path, _ := range mimetypeMap {
		paths = append(paths, path)
	}
	return paths
}

func Run() {
	fmt.Println("Ind i applications.Run")
	var err error
	watcher, err = fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	defer watcher.Close()

	for _, dataDir := range append(xdg.DataDirs, xdg.DataHome) {
		watchDir(dataDir + "/applications")
	}
	watchDir(xdg.ConfigHome)

	collectAndMap()
	for {
		select {
		case event := <-watcher.Events:
			fmt.Println("Event:", event)
			if isRelevant(event) {
				collectAndMap()
			}
		case err := <-watcher.Errors:
			fmt.Println(err)
		}
	}
}

var watcher *fsnotify.Watcher

func watchDir(dir string) {
	if xdg.DirOrFileExists(dir) {
		if err := watcher.Add(dir); err != nil {
			fmt.Println("Could not watch:", dir, ":", err)
		}
	}
}

func isRelevant(event fsnotify.Event) bool {
	return strings.HasSuffix(event.Name, ".desktop") || strings.HasSuffix(event.Name, "/mimeapps.list")
}

var mimetypes, applications, actions atomic.Value

func init() {
	mimetypes.Store(make(MimetypeMap))
	applications.Store(make(ApplicationMap))
	actions.Store(make(actionMap))
}

func collectAndMap() {
	fmt.Println("collect mimetypes and applications")
	var c = Collect()
	var mimetypeMap = make(MimetypeMap, len(c.mimetypes))
	for mimeId, mimetype := range c.mimetypes {
		mimetypeMap["/mimetype/"+mimeId] = mimetype
	}
	mimetypes.Store(mimetypeMap)

	var appMap = make(ApplicationMap, len(c.applications))
	var actMap = make(actionMap, 50)
	for appId, app := range c.applications {
		appMap[appSelf(appId)] = app
		if len(app.DesktopActions) > 0 {
			for _, act := range app.DesktopActions {
				fmt.Println("Mapping action", act.self)
				actMap[act.self] = act
			}
		}
	}
	applications.Store(appMap)
	actions.Store(actMap)
}

/*if otherActionsPath != "" {
	for daId, da := range app.DesktopActions {
		var exec, terminal = da.Exec, app.Terminal
		resources = append(resources, &server.JsonData{
			Self:        appSelf(appId) + "/" + daId,
			Type:        "Action",
			Title:       da.Name,
			IconName:    da.IconName,
			OnPost:      da.Name,
			PostHandler: func(w http.ResponseWriter, r *http.Request) { launch(exec, terminal) },
		})
	}
}*/
