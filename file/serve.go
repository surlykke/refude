// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package file

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if f := GetResource(path.Of(r.URL.Path)); f == nil {
		respond.NotFound(w)
	} else {
		resource.ServeSingleResource(w, r, f)
	}
}

func GetResource(resPath path.Path) *File {
	var pathS = string(resPath)
	fmt.Println("file.GetResource, path: '" + resPath + "'")
	if !strings.HasPrefix(pathS, "/file/") {
		log.Warn("Unexpeded path:", resPath)
		return nil
	} else if file, err := makeFileFromPath(pathS[5:]); err != nil {
		log.Warn("Could not make file from", pathS[5:], err)
		return nil
	} else if file == nil {
		fmt.Println(".. not found")
		return nil
	} else {
		return file
	}
}

var watchedDirs []string
var files []*File = []*File{}
var filesLock sync.Mutex

func Run() {

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
	scanDirs()
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
			scanDirs()
		}
	}

}

func scanDirs() {
	var collected = make([]*File, 0, 50)
	for _, dir := range watchedDirs {
		for _, entry := range readEntries(dir) {
			var name = entry.Name()
			if file, err := makeFileFromPath(dir + "/" + name); err == nil {
				collected = append(collected, file)
			}
		}
	}
	filesLock.Lock()
	files = collected
	filesLock.Unlock()
}

func GetFiles() []resource.Resource {
	filesLock.Lock()
	defer filesLock.Unlock()
	var result = make([]resource.Resource, len(files), len(files))
	for i, file := range files {
		result[i] = file
	}
	return result
}
