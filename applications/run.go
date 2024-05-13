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

type AppData struct {
	DesktopId string
	Title     string
	IconUrl   string
}

type openFileRequest struct {
	appId string
	path  string
}

// f.AddLink("?action="+app.DesktopId, "Open with "+app.Title, app.IconUrl, relation.Action)

type MimetypeAppDataChan chan map[string][]AppData

var mimetypeAppDataChans []MimetypeAppDataChan

func MakeMimetypeAppDataChan() MimetypeAppDataChan {
	mimetypeAppDataChans = append(mimetypeAppDataChans, make(MimetypeAppDataChan))
	return mimetypeAppDataChans[len(mimetypeAppDataChans)-1]
}

type AppIdAppDataChan chan map[string]AppData

var appIdAppDataChans []AppIdAppDataChan

func MakeAppdIdAppDataChan() AppIdAppDataChan {
	appIdAppDataChans = append(appIdAppDataChans, make(AppIdAppDataChan))
	return appIdAppDataChans[len(appIdAppDataChans)-1]
}

var foundApps = make(chan map[string]*DesktopApplication)
var foundMimetypes = make(chan map[string]*Mimetype)
var appRequests = repo.MakeAndRegisterRequestChan()
var appRepo = repo.MakeRepoWithFilter[*DesktopApplication](filter)
var mtRequests = repo.MakeAndRegisterRequestChan()
var mimetypeRepo = repo.MakeRepo[*Mimetype]()

var openFileRequests = make(chan openFileRequest)

func Run() {
	go watch()
	go appLoop()
	go mtLoop()
}

func OpenFile(appId, path string) {
	if appId == "" {
		xdg.RunCmd("xdg-open", path)
	} else {
		openFileRequests <- openFileRequest{appId: appId, path: path}
	}
}

func appLoop() {
	for {
		select {
		case apps := <-foundApps:
			appRepo.RemoveAll()
			for _, app := range apps {
				appRepo.Put(app)
			}
		case req := <-appRequests:
			appRepo.DoRequest(req)
		case ofr := <-openFileRequests:
			if da, ok := appRepo.Get("/application/" + ofr.appId); ok {
				da.Run(ofr.path)
			}
		}
	}
}

func mtLoop() {
	for {
		select {
		case mimetypes := <-foundMimetypes:
			mimetypeRepo.RemoveAll()
			for _, mt := range mimetypes {
				mimetypeRepo.Put(mt)
			}
		case req := <-mtRequests:
			mimetypeRepo.DoRequest(req)
		}
	}
}

func watch() {
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

	for event := range watcher.Events {
		if isRelevant(event) {
			Collect()
		}
	}

}

func filter(term string, app *DesktopApplication) bool {
	return len(term) > 0 && !app.Hidden
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
