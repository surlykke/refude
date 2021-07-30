// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func Run() {
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

	collectionStore.Store(Collect())
	for {
		select {
		case event := <-watcher.Events:
			if isRelevant(event) {
				collectionStore.Store(Collect())
			}
		case err := <-watcher.Errors:
			log.Warn(err)
		}
	}
}

var watcher *fsnotify.Watcher

func watchDir(dir string) {
	if xdg.DirOrFileExists(dir) {
		if err := watcher.Add(dir); err != nil {
			log.Warn("Could not watch:", dir, ":", err)
		}
	}
}

func isRelevant(event fsnotify.Event) bool {
	return strings.HasSuffix(event.Name, ".desktop") || strings.HasSuffix(event.Name, "/mimeapps.list")
}

var collectionStore atomic.Value

func init() {
	collectionStore.Store(makeCollection())
}

func GetMtList(mimetypeId string) []string {
	var mimetypes = collectionStore.Load().(collection).mimetypes
	var result = make([]string, 1, 5)
	result[0] = mimetypeId
	for i := 0; i < len(result); i++ {
		if mt, ok := mimetypes[result[i]]; ok {
			for _, super := range mt.SubClassOf {
				result = slice.AppendIfNotThere(result, super)
			}
		}
	}
	return result
}

func GetApp(appId string) *DesktopApplication {
	return collectionStore.Load().(collection).applications[appId]
}

func GetAppsForMimetype(mimetypeId string) (recommended, other []*DesktopApplication) {
	var c = collectionStore.Load().(collection)

	recommended = make([]*DesktopApplication, 0, 10)
	other = make([]*DesktopApplication, 0, len(c.applications))

	for _, app := range c.applications {
		if argPlaceholders.MatchString(app.Exec) {
			for _, mt := range app.Mimetypes {
				if mt == mimetypeId {
					recommended = append(recommended, app)
					goto next
				}
			}
			other = append(other, app)
		next:
		}
	}

	return recommended, other
}
