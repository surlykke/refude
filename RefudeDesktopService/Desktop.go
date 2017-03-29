/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (

	"os"
	"path/filepath"
	"strings"
	"fmt"
	"github.com/surlykke/RefudeServices/xdg"
	"github.com/surlykke/RefudeServices/common"
	"github.com/surlykke/RefudeServices/service"
	"golang.org/x/sys/unix"
	"net/http"
	"time"
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
)


var applications = make(map[string]*DesktopApplication)
var mimetypes = make(map[string]*Mimetype)
var actionPaths = make(common.StringList, 0)

var historyPath = xdg.DataHome() + "/data/RefudeDesktopServiceHistory"

var Activations = make(chan string)
var fileChanges = make(chan bool)

var history = make(map[string]time.Time)

func DesktopRun() {

	getHistory()
	update()
	go watchFiles()
	for {
		select {
		case url:= <-Activations:
			history[url] = time.Now()
			sortAndMapActionPaths()
			saveHistory()
		case <-fileChanges:
			update()
		}
	}
}

func watchFiles() {
	fd, err := unix.InotifyInit()
	defer unix.Close(fd)

	if err != nil {
		panic(err)
	}
	for _,dataDir := range append(xdg.DataDirs(), xdg.DataHome()) {
		appDir := dataDir + "/applications"
		fmt.Println("Watching: " + appDir)
		_, err := unix.InotifyAddWatch(fd, appDir, unix.IN_CREATE | unix.IN_MODIFY | unix.IN_DELETE)
		if  err != nil {
			panic(err)
		}
	}

	dummy := make([]byte, 100)
	for {
		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fileChanges <- true
	}
}



func getHistory() {
	if file, err := os.Open(historyPath); err == nil {
		defer file.Close()
		if bytes, err := ioutil.ReadAll(file); err == nil {
			json.Unmarshal(bytes, &history)
		} else {
			fmt.Println("Error unmarshalling history ", err)
		}
	} else {
		fmt.Println("Error opening file ", historyPath, " ", err)
	}
}

func saveHistory() {
	if bytes, err := json.Marshal(history); err == nil {
		if file, err := os.Create(historyPath); err == nil {
			defer file.Close()
			file.Write(bytes)
		} else {
			fmt.Println("Error opening ", historyPath, " ", err)
		}
	} else {
		log.Println("Error marshalling history:", err)
	}
}

func update() {
	c := NewCollector()
	c.collect()

	for applicationId,application := range applications {
		if _,ok := c.applications[applicationId]; !ok {
			service.Unmap("/application/" + applicationId)
			for actionId := range application.Actions {
				service.Unmap("/action/" + applicationId + "_" + actionId)
			}
		}
	}

	applications = c.applications
	var applicationPaths = make(common.StringList, 0)
	for applicationId, application := range applications {
		service.Map("/application/" + applicationId, application)
		applicationPaths = append(applicationPaths, "application/" + applicationId)
		for actionId, action := range application.Actions {
			service.Map("/action/" + applicationId + "_" + actionId, action)
			actionPaths = append(actionPaths, "action/" + applicationId + "_" + actionId)
		}
	}
	service.Map("/applications", applicationPaths)
	sortAndMapActionPaths()

	for mimetypeId := range mimetypes {
		if _,ok := c.mimetypes[mimetypeId]; !ok {
			service.Unmap("/mimetype/" + mimetypeId)
		}
	}

	mimetypes = c.mimetypes
	var mimetypePaths = make(common.StringList, 0)
	for mimetypeId, mimeType := range mimetypes {
		service.Map("/mimetype/" + mimetypeId, mimeType)
		mimetypePaths = append(mimetypePaths, "mimetype/" + mimetypeId)
	}
	service.Map("/mimetypes", mimetypePaths)
}

func sortAndMapActionPaths() {
	sort.SliceStable(actionPaths, func(i int, j int) bool {
		p1 := actionPaths[i]
		p2 := actionPaths[j]
		t1,_ := history[p1]
		t2,_ := history[p2]
		return t1.After(t2)
	})

	service.Map("/actions", actionPaths)
}

type Collector struct {
	applications map[string]*DesktopApplication
	mimetypes map[string]*Mimetype
	actions map[string]*Action
}

func NewCollector() Collector {
	return Collector{make(map[string]*DesktopApplication), make(map[string]*Mimetype), make(map[string]*Action)}
}

func (c *Collector) collect()  {
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)

	for _, dir := range xdg.DataDirs() {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome() + "/applications")

	for _,dir := range append(xdg.ConfigDirs(), xdg.ConfigHome()) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for _,app := range c.applications {
		for key,action := range app.Actions {
			actionId := app.Id + "_" + key
			c.actions[actionId] = action
		}
	}
}

func (c* Collector) addAssociations(mimeId string, appIds...string) {
	if mimetype, mimetypeFound := c.mimetypes[mimeId]; mimetypeFound {
		for _,appId := range appIds {
			if application, appFound := c.applications[appId]; appFound {
				mimetype.AssociatedApplications = common.AppendIfNotThere(mimetype.AssociatedApplications, appId)
				application.Mimetypes = common.AppendIfNotThere(application.Mimetypes, mimeId)
			}
		}
	}
}

func (c* Collector) removeAssociations(mimeId string, appIds...string) {
	mimetype, mimetypeFound := c.mimetypes[mimeId]

	for _,appId := range appIds {
		if app, ok := c.applications[appId]; ok {
			app.Mimetypes = common.Remove(app.Mimetypes, mimeId)
		}
		if mimetypeFound {
			mimetype.AssociatedApplications = common.Remove(mimetype.AssociatedApplications, appId)
		}
	}
}

func (c *Collector) collectApplications(appdir string) {
	filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
		if !(info.IsDir() || !strings.HasSuffix(path, ".desktop")) {
			app, mimetypes, err := readDesktopFile(path)
			if err == nil {
				app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
				if app.Hidden {
					delete(c.applications, app.Id)
				} else {
					c.applications[app.Id] = app
					for _, mimetypeId := range mimetypes {
						c.addAssociations(mimetypeId, app.Id)
					}
				}
			}
		}
		return nil
	})

}

func (c *Collector) readMimeappsList(path string) {
    mimeappsList := struct {
		defaultApplications map[string][]string
		addedAssociations   map[string][]string
		removedAssociations map[string][]string
	}{make(map[string][]string), make(map[string][]string), make(map[string][]string)}

	iniGroups, err := common.ReadIniFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
		}
		mimeappsList.addedAssociations = make(map[string][]string)
		mimeappsList.removedAssociations = make(map[string][]string)
		mimeappsList.defaultApplications = make(map[string][]string)
	} else {
		for _, iniGroup := range iniGroups {
			var dest map[string][]string
			switch iniGroup.Name {
			case "Default Applications":
				dest = mimeappsList.defaultApplications
			case "Added Associations":
				dest = mimeappsList.addedAssociations
			case "Removed Associations":
				dest = mimeappsList.removedAssociations
			default:
				continue
			}
			for k, v := range iniGroup.Entry {
				dest[k] = common.Split(v, ";")
			}
		}
	}

	for mimeId, appIds := range mimeappsList.removedAssociations {
		c.removeAssociations(mimeId, appIds...)
	}

	for mimeId, appIds := range mimeappsList.addedAssociations {
		c.addAssociations(mimeId, appIds...)
	}

	for mimetypeId, appIds := range mimeappsList.defaultApplications {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _,appId := range appIds {
				mimetype.DefaultApplications = append(mimetype.DefaultApplications, appId)
			}
		}
	}
}

type Icon struct {
	prefix string
}

func (i Icon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, i.prefix) {
		http.ServeFile(w, r, r.URL.Path[len(i.prefix):])
	}
}
