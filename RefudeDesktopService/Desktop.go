// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"strings"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/lib/service"
	"golang.org/x/sys/unix"
	"net/http"
	"regexp"
	"github.com/surlykke/RefudeServices/lib/resource"
	"time"
)


var fileChange = make(chan string)
var launch = make(chan string)

func Run() {
	LoadLastLaunched()
	go WatchFiles()
	update()
	for {
		select {
		case appId := <-launch:
			path := "/applications/" + appId
			fmt.Println("Launch: ", path)
			if res := service.Get(path); res != nil {
				desktopApplication := res.Data.(*DesktopApplication).Copy()
				desktopApplication.RelevanceHint = time.Now().UnixNano()/1000000
				service.Map(path, resource.JsonResource(desktopApplication, DesktopApplicationPOST))
				lastLaunched[appId] = desktopApplication.RelevanceHint
				SaveLastLaunched()
			}
		case <-fileChange:
			update()
		}
	}

}

func WatchFiles() {

	fd, err := unix.InotifyInit()
	defer unix.Close(fd)

	if err != nil {
		panic(err)
	}
	for _,dataDir := range append(xdg.DataDirs, xdg.DataHome) {
		appDir := dataDir + "/applications"
		fmt.Println("Watching: " + appDir)
		if _, err := unix.InotifyAddWatch(fd, appDir, unix.IN_CREATE | unix.IN_MODIFY | unix.IN_DELETE); err != nil {
			panic(err)
		}
	}

	if _, err := unix.InotifyAddWatch(fd, xdg.ConfigHome + "/mimeapps.list", unix.IN_CLOSE_WRITE); err != nil {
		panic(err)
	}

	dummy := make([]byte, 100)
	for {
		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Something happened...")
		fileChange <- ""
	}
}


var applicationIds = make([]string, 0)
var mimetypeIds = make([]string, 0)

func update() {
	c := Collect()

	for _, appId := range applicationIds {
		if _,ok := c.applications[appId]; !ok {
			service.Unmap("/applications/" + appId)
		}
	}

	for appId, newDesktopApplication := range c.applications {
		newDesktopApplication.RelevanceHint = lastLaunched[newDesktopApplication.Id]
		service.Map("/applications/" + appId, resource.JsonResource(newDesktopApplication, DesktopApplicationPOST))
	}

	for _, mimetypeId := range mimetypeIds {
		if _,ok := c.mimetypes[mimetypeId]; !ok {
			service.Unmap("/mimetypes/" + mimetypeId)
		}
	}

	for mimetypeId, mimeType := range c.mimetypes {
		service.Map("/mimetypes/" + mimetypeId, resource.JsonResource(mimeType, MimetypePOST))
	}

}

func IconGet(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	prefix := this.Data.(string)
	if strings.HasPrefix(r.URL.Path, prefix) {
		http.ServeFile(w, r, r.URL.Path[len(prefix):])
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

var mimetypePathPattern = func() *regexp.Regexp {
	if pattern, err := regexp.Compile(`/mimetypes/[^/]+/[^/]+`); err != nil {
		panic(err)
	} else {
		return pattern
	}
}()

type MimetypePostPayload struct {
	DefaultApplication string
}



func RequestInterceptor(w http.ResponseWriter, r* http.Request) {
	if strings.HasPrefix(r.URL.Path, "/mimetypes/x-scheme-handler/") && ! service.Has(r.URL.Path) {
		mimetypeId := r.URL.Path[len("/mimetypes/"):]
		if mimetype, err := NewMimetype(mimetypeId); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			fmt.Println("Mapping ", mimetype)
			service.Map(r.URL.Path, resource.JsonResource(mimetype, MimetypePOST))
		}
	}

	service.ServeHTTP(w, r)
}


