// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var FileMap = repo.MakeSynkMap[string, *File]()

func Run() {
	var watchedDirs []string

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, dir := range []string{xdg.Home, xdg.DesktopDir, xdg.DownloadDir, xdg.TemplatesDir, xdg.PublicshareDir, xdg.DocumentsDir, xdg.MusicDir, xdg.PicturesDir, xdg.VideosDir} {
		if xdg.DirOrFileExists(dir) {
			if err := watcher.Add(dir); err != nil {
				log.Warn("Not watching", dir, err)
			} else {
				watchedDirs = append(watchedDirs, dir)
			}
		}
	}

	var scanEv = make(chan struct{})
	var scanScheduled = false
	scanDirs(watchedDirs)
	go func() {
		var appSubscription = applications.AppEvents.Subscribe()
		for {
			scanEv <- appSubscription.Next()

		}
	}()

	for {
		select {
		case <-watcher.Events:
			// We collect events for a second before scanning, as they can come in bursts
			if !scanScheduled {
				scanScheduled = true
				go func() { time.Sleep(1 * time.Second); scanEv <- struct{}{} }()
			}
		case <-scanEv:
			scanScheduled = false
			scanDirs(watchedDirs)
		}
	}

}

func scanDirs(watchedDirs []string) {
	var collected = make(map[string]*File, 50)
	for _, dir := range watchedDirs {
		for _, entry := range readEntries(dir) {
			var name = entry.Name()
			var path = dir + "/" + name
			if file, err := makeFileFromPath(path); err == nil {
				collected[path[1:]] = file
			}
		}
	}
	FileMap.Replace(collected)
}
