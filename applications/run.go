// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/resourcerepo"
	"github.com/surlykke/RefudeServices/lib/searchutils"
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

	Collect()
	for {
		select {
		case event := <-watcher.Events:
			if isRelevant(event) {
				Collect()
			}
		case err := <-watcher.Errors:
			log.Warn(err)
		}
	}
}

func Search(list *resource.RRList, term string) {
	if len(term) > 0 {
		for _, da := range resourcerepo.GetTypedByPrefix[*DesktopApplication]("/application/") {
			if rnk := searchutils.Match(term, da.Title, da.Keywords...); rnk >= 0 {
				if !da.Hidden {
					*list = append(*list, resource.RankedResource{Res: da, Rank: rnk})
				}
			}
		}
	}
}

var (
	listeners    []func()
	listenerLock sync.Mutex
)

func AddListener(listener func()) {
	listenerLock.Lock()
	defer listenerLock.Unlock()
	listeners = append(listeners, listener)
}

func notifyListeners() {
	listenerLock.Lock()
	defer listenerLock.Unlock()
	for _, listener := range listeners {
		listener()
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
