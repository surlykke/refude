// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"log"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/refude/internal/applications"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/xdg"
)

var FileMap = entity.MakeMap[string, *File]()

func Run() {
	var watchedDirs []string

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, dir := range []string{xdg.Home, xdg.DesktopDir, xdg.DownloadDir, xdg.TemplatesDir, xdg.PublicshareDir, xdg.DocumentsDir, xdg.MusicDir, xdg.PicturesDir, xdg.VideosDir} {
		if xdg.DirOrFileExists(dir) {
			if err := watcher.Add(dir); err != nil {
				log.Print("Not watching", dir, err)
			} else {
				watchedDirs = append(watchedDirs, dir)
			}
		}
	}

	var scanEv = make(chan struct{})
	var scanScheduled = false
	scanDirs(watchedDirs)
	go func() {
		var appSubscription = applications.AppMap.Events.Subscribe()
		for {
			appSubscription.Next()
			scanEv <- struct{}{}
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
			if file, err := makeFileFromPath(path); err == nil && file != nil {
				collected[path[1:]] = file
			}
		}
	}
	FileMap.ReplaceAll(collected)
}
