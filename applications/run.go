// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var subscribtions = make([]chan Collection, 0, 10)

func SubscribeToCollections() chan Collection {
	var subscription = make(chan Collection)
	subscribtions = append(subscribtions, subscription)
	return subscription
}

var launchRequests = make(chan []string)

func Launch(appId string, args ...string) {
	var s = []string{appId}
	s = append(s, args...)
	launchRequests <- s
}

func Run() {
	var appMaps = make(chan map[string]*DesktopApplication)
	go runAppRepo(appMaps)

	var mtMaps = make(chan map[string]*Mimetype)
	go runMimetypeRepo(mtMaps)

	var desktopFileEvents = make(chan struct{})
	go watchForDesktopFiles(desktopFileEvents)

	for {
		var collection Collection = collect()
		appMaps <- collection.Apps
		mtMaps <- collection.Mimetypes
		for _, subscription := range subscribtions {
			subscription <- collection
		}
		<-desktopFileEvents
	}

}

func runAppRepo(appMaps chan map[string]*DesktopApplication) {
	var appRepo = repo.MakeRepoWithFilter(filter)
	var appRequests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case appMap := <-appMaps:
			appRepo.RemoveAll()
			for _, app := range appMap {
				appRepo.Put(app)
			}
		case appRequest := <-appRequests:
			appRepo.DoRequest(appRequest)
		case req := <-launchRequests:
			if len(req) > 0 {
				var path = "/application/" + req[0]
				if da, ok := appRepo.Get(path); ok {
					da.Run(strings.Join(req[1:], " "))
				}
			}
		
		}

	}
}

func runMimetypeRepo(mimetypeMaps chan map[string]*Mimetype) {
	var mimetypeRepo = repo.MakeRepo[*Mimetype]()
	var requests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case mimetypeMap := <-mimetypeMaps:
			mimetypeRepo.RemoveAll()
			for _, mt := range mimetypeMap {
				mimetypeRepo.Put(mt)
			}
		case req := <-requests:
			mimetypeRepo.DoRequest(req)
		}

	}
}

func filter(term string, app *DesktopApplication) bool {
	return len(term) > 0 && !app.Hidden
}

func watchForDesktopFiles(events chan struct{}) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, dir := range append(xdg.DataDirs, xdg.DataHome) {
		if xdg.DirOrFileExists(dir) {
			if err := watcher.Add(dir); err != nil {
				log.Warn("Could not watch:", dir, ":", err)
			}
		}
	}

	for event := range watcher.Events {
		if strings.HasSuffix(event.Name, ".desktop") {
			events <- struct{}{}
		}
	}

}

