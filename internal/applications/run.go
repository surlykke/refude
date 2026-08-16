// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package applications

import (
	"log"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/xdg"
	"github.com/surlykke/refude/pkg/pubsub"
)

var AppEvents = pubsub.MakePublisher[struct{}]()

func Run() {
	var desktopFileEvents = make(chan struct{})
	go watchForDesktopFiles(desktopFileEvents)

	for {
		var apps, mimetypes = collect()
		resource.Replace(slices.Collect(maps.Values(apps)), "/application/")
		resource.Replace(slices.Collect(maps.Values(mimetypes)), "/mimetype/")
		AppEvents.Publish(struct{}{})
		<-desktopFileEvents
	}

}

func GetHandlers(mimetype string) []DesktopApplication {
	var apps = make([]DesktopApplication, 0, 10)
	if mt, ok := resource.Get[Mimetype](mimetypePathPrefix + mimetype); ok {
		for _, appId := range mt.Applications {
			if app, ok := resource.Get[DesktopApplication](appPathPrefix + appId); ok {
				apps = append(apps, app)
			}
		}
	}
	return apps
}

func GetTitleAndIcon(appId string) (string, string, bool) {
	if appId != "" {
		if app, ok := resource.Get[DesktopApplication](appPathPrefix + appId); ok {
			return app.Title, app.Icon, true
		}
	}
	return "", "", false
}

func OpenFile(appId, filePath string) bool {
	if app, ok := resource.Get[DesktopApplication](appPathPrefix + appId); ok {
		app.Run(filePath)
		return true
	}
	return false
}

func watchForDesktopFiles(events chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	var filesToWatch = make([]string, 0, 20)
	for _, s := range append(xdg.DataDirs, xdg.DataHome) {
		filesToWatch = append(filesToWatch, s+"/applications")
	}
	filesToWatch = append(filesToWatch, xdg.ConfigHome+"/mimeapps.list")

	for _, f := range filesToWatch {
		if xdg.DirOrFileExists(f) {
			if err := watcher.Add(f); err != nil {
				log.Print("Could not watch:", f, ":", err)
			}
		}
	}

	var gracePeriodEnded = make(chan struct{})
	var reloadScheduled = false
	for {
		select {
		// When the user reinstalls something it will create a number of inotify events. We collect for a couple of seconds
		// before doing a reload.
		case event := <-watcher.Events:
			if !reloadScheduled && strings.HasSuffix(event.Name, ".desktop") || strings.HasSuffix(event.Name, "/mimeapps.list") {
				reloadScheduled = true
				go func() {
					time.Sleep(2 * time.Second)
					gracePeriodEnded <- struct{}{}
				}()
			}
		case _ = <-gracePeriodEnded:
			reloadScheduled = false
			events <- struct{}{}
		}
	}
}
