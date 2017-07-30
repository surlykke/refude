// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/sys/unix"
)

var fileChange = make(chan string)
var launch = make(chan string)

func Run() {
	LoadLastLaunched()

	fd, err := unix.InotifyInit()
	defer unix.Close(fd)

	if err != nil {
		panic(err)
	}
	for _, dataDir := range append(xdg.DataDirs, xdg.DataHome) {
		appDir := dataDir + "/applications"
		fmt.Println("Watching: " + appDir)
		if _, err := unix.InotifyAddWatch(fd, appDir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE); err != nil {
			panic(err)
		}
	}

	if _, err := unix.InotifyAddWatch(fd, xdg.ConfigHome+"/mimeapps.list", unix.IN_CLOSE_WRITE); err != nil {
		panic(err)
	}

	update()
	dummy := make([]byte, 100)
	for {
		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Something happened...")
		update()
	}
}

var applicationIds = make([]string, 0)
var mimetypeIds = make([]string, 0)

func update() {
	c := Collect()

	for _, appId := range applicationIds {
		if _, ok := c.applications[appId]; !ok {
			service.Unmap("/applications/" + appId)
		}
	}

	for appId, newDesktopApplication := range c.applications {
		newDesktopApplication.RelevanceHint = lastLaunched[newDesktopApplication.Id]
		service.Map("/applications/"+appId, newDesktopApplication)
		if newDesktopApplication.IconUrl != "" {
			iconPath := IconPath(newDesktopApplication.IconPath)
			urlPath := string("/icons" + iconPath)
			service.Map(urlPath, iconPath)
		}
		for actionId, action := range newDesktopApplication.Actions {
			if actionId != "_default" && action.IconUrl != "" {
				iconPath := IconPath(action.IconPath)
				urlPath := string("/icons" + iconPath)
				service.Map(urlPath, iconPath)
			}
		}
	}

	for _, mimetypeId := range mimetypeIds {
		if _, ok := c.mimetypes[mimetypeId]; !ok {
			service.Unmap("/mimetypes/" + mimetypeId)
		}
	}

	for mimetypeId, mimeType := range c.mimetypes {
		service.Map("/mimetypes/"+mimetypeId, mimeType)
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

func RequestInterceptor(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/mimetypes/x-scheme-handler/") && !service.Has(r.URL.Path) {
		mimetypeId := r.URL.Path[len("/mimetypes/"):]
		if mimetype, err := NewMimetype(mimetypeId); err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			fmt.Println("Mapping ", mimetype)
			service.Map(r.URL.Path, mimetype)
		}
	}

	service.ServeHTTP(w, r)
}
