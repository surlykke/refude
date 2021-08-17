// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
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
