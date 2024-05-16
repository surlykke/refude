// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"fmt"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type AppSummary struct {
	DesktopId string
	Title     string
	IconUrl   string
}

var appSummarySubscribers []chan []AppSummary
var mimetypeHandlerSubscribers []chan map[string][]string
var launchRequests = make(chan []string)

func SubscribeToAppSummary() chan []AppSummary {
	appSummarySubscribers = append(appSummarySubscribers, make(chan []AppSummary))
	return appSummarySubscribers[len(appSummarySubscribers)-1]
}

func SubscribeToMimetypeHandlers() chan map[string][]string {
	mimetypeHandlerSubscribers = append(mimetypeHandlerSubscribers, make(chan map[string][]string))
	return mimetypeHandlerSubscribers[len(mimetypeHandlerSubscribers)-1]
}

func Launch(appId string, args ...string) {
	var s = []string{appId}
	s = append(s, args...)
	launchRequests <- s
}

var appRequests = repo.MakeAndRegisterRequestChan()
var appRepo = repo.MakeRepoWithFilter[*DesktopApplication](filter)

func Run() {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	for _, dataDir := range append(xdg.DataDirs, xdg.DataHome) {
		watchDir(watcher, dataDir+"/applications")
	}
	watchDir(watcher, xdg.ConfigHome)

	collect()

	for {
		select {
		case event := <-watcher.Events:
			if strings.HasSuffix(event.Name, ".desktop") {
				collect()
			}
		case req := <-appRequests:
			appRepo.DoRequest(req)
		case req := <-launchRequests:
			fmt.Println("Launch request:", req)
			if len(req) > 0 {
				var path = "/application/" + req[0]
				fmt.Println("Looking for", path)
				if da, ok := appRepo.Get(path); ok {
					fmt.Println("Try to run ", strings.Join(req[1:], " "))
					da.Run(strings.Join(req[1:], " "))
				}
			}
		}

	}
}

func filter(term string, app *DesktopApplication) bool {
	return len(term) > 0 && !app.Hidden
}

func watchDir(watcher *fsnotify.Watcher, dir string) {
	if xdg.DirOrFileExists(dir) {
		if err := watcher.Add(dir); err != nil {
			log.Warn("Could not watch:", dir, ":", err)
		}
	}
}
