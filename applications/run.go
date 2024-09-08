// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"fmt"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/pubsub"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var AppEvents = pubsub.MakePublisher[struct{}]()

func Run() {
	var desktopFileEvents = make(chan struct{})
	go watchForDesktopFiles(desktopFileEvents)

	for {
		var collection Collection = collect()

		var apps = make([]resource.Resource, 0, len(collection.Apps))
		for _, app := range collection.Apps {
			apps = append(apps, app)
		}
		repo.Replace(apps, "/application/")

		var mts = make([]resource.Resource, 0, len(collection.Mimetypes))
		for _, mt := range collection.Mimetypes {
			mts = append(mts, mt)
		}
		repo.Replace(mts, "/mimetype/")

		AppEvents.Publish(struct{}{})

		<-desktopFileEvents
	}

}

func GetHandlers(mimetype string) []*DesktopApplication {
	var apps = make([]*DesktopApplication, 0, 10)
	if mt, ok := repo.Get[*Mimetype]("/mimetype/" + mimetype); ok {
		for _, appId := range mt.Applications {
			if app, ok := repo.Get[*DesktopApplication]("/application/" + appId); ok {
				apps = append(apps, app)
			}
		}
	}
	return apps
}

func GetIconUrl(appId string) string {
	if da, ok := repo.Get[*DesktopApplication]("/application/" + appId); ok {
		return da.GetLinks().Get(relation.Icon).Href
	}
	return ""
}

func OpenFile(appId, path string) bool {
	fmt.Println("Openfile looking for", "/application/"+appId)
	if app, ok := repo.Get[*DesktopApplication]("/application/" + appId); ok {
		app.Run(path)
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
				log.Warn("Could not watch:", f, ":", err)
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
			fmt.Println("file event:", event.Name)
			if !reloadScheduled && strings.HasSuffix(event.Name, ".desktop") || strings.HasSuffix(event.Name, "/mimeapps.list") {
				reloadScheduled = true
				go func() {
					fmt.Println("Gracing...")
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
