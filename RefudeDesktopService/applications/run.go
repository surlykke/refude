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

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func AppServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/applications" {
		respond.AsJson(w, r, CollectApps(searchutils.Term(r)))
	} else if app, ok := applications.Load().(ApplicationMap)[r.URL.Path]; ok {
		if r.Method == "POST" {
			respond.AcceptedAndThen(w, func() { launch(app.Exec, app.Terminal) })
		} else {
			respond.AsJson(w, r, app.ToStandardFormat())
		}
	} else if app = getAppForActions(r.URL.Path); app != nil {
		respond.AsJson(w, r, app.collectActions(searchutils.Term(r)))
	} else if act := getDesktopAction(r.URL.Path); act != nil {
		if r.Method == "POST" {
			respond.AcceptedAndThen(w, func() { launch(act.Exec, false) })
		} else {
			respond.AsJson(w, r, act.ToStandardFormat())
		}
	} else {
		respond.NotFound(w)
	}
}

func getAppForActions(path string) *DesktopApplication {
	if strings.HasSuffix(path, "/actions") {
		if app, ok := applications.Load().(ApplicationMap)[path[:len(path)-8]]; ok {
			if len(app.DesktopActions) > 0 {
				return app
			}
		}
	}
	return nil
}

func getDesktopAction(path string) *DesktopAction {
	if i := strings.Index(path, "/action/"); i > -1 {
		if app, ok := applications.Load().(ApplicationMap)[path[:i]]; ok {
			if act, ok := app.DesktopActions[path[i+8:]]; ok {
				return act
			}
		}
	}

	return nil
}

func CollectApps(term string) respond.StandardFormatList {
	var appMap = applications.Load().(ApplicationMap)
	var sfl = make(respond.StandardFormatList, 0, len(appMap))
	for _, app := range appMap {
		if rank := searchutils.SimpleRank(app.Name, app.Comment, term); rank > -1 {
			sfl = append(sfl, app.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

func MimetypeServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/mimetypes" {
		respond.AsJson(w, r, CollectMimetypes(searchutils.Term(r)))
	} else if mt, ok := mimetypes.Load().(MimetypeMap)[r.URL.Path]; ok {
		respond.AsJson(w, r, mt.ToStandardFormat())
	} else {
		respond.NotFound(w)
	}
}

func CollectMimetypes(term string) respond.StandardFormatList {
	var mimeMap = mimetypes.Load().(MimetypeMap)
	var sfl = make(respond.StandardFormatList, 0, len(mimeMap))
	for _, mt := range mimeMap {
		if rank := searchutils.SimpleRank(mt.Comment, mt.Acronym, term); rank > -1 {
			sfl = append(sfl, mt.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

func AllPaths() []string {
	var applicationMap = applications.Load().(ApplicationMap)
	var mimeMap = mimetypes.Load().(MimetypeMap)
	var paths = make([]string, 0, len(applicationMap)+len(mimeMap)+100)
	for path, app := range applicationMap {
		paths = append(paths, path)
		if len(app.DesktopActions) > 0 {
			paths = append(paths, path+"/actions")
			for _, act := range app.DesktopActions {
				paths = append(paths, act.self)
			}
		}
	}
	for path, _ := range mimeMap {
		paths = append(paths, path)
	}
	paths = append(paths, "/applications", "/mimetypes")
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

var mimetypes, applications atomic.Value

func init() {
	mimetypes.Store(make(MimetypeMap))
	applications.Store(make(ApplicationMap))
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
	for appId, app := range c.applications {
		appMap[appSelf(appId)] = app
	}
	applications.Store(appMap)
}
