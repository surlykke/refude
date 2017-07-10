// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (

	"os"
	"path/filepath"
	"strings"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/lib/service"
	"golang.org/x/sys/unix"
	"net/http"
	"regexp"
	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/resource"
)


var applicationIds = make([]string, 0)
var mimetypeIds = make([]string, 0)


func DesktopRun() {

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

func update() {
	c := NewCollector()
	c.collect()


	for _, appId := range applicationIds {
		if _,ok := c.applications[appId]; !ok {
			service.Unmap("/applications/" + appId)
		}
	}

	for appId, newDesktopApplication := range c.applications {
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


type Collector struct {
	applications map[string]*DesktopApplication
	mimetypes map[string]*Mimetype
}

func NewCollector() Collector {
	return Collector{make(map[string]*DesktopApplication), make(map[string]*Mimetype)}
}

func (c *Collector) collect()  {
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)

	for _, dir := range xdg.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome + "/applications")

	for _,dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}
}

func (c *Collector) getMimetype(id string) *Mimetype {
	if mimetype, ok := c.mimetypes[id]; ok {
		return mimetype
	} else if mimetype, err := NewMimetype(id); err == nil {
		c.mimetypes[id] = mimetype
		return mimetype
	} else {
		return nil
	}
}

func appPath(id string) string {
	return "/application/" + id
}


func (c* Collector) addAssociations(mimeId string, appIds...string) {
	if mimetype := c.getMimetype(mimeId); mimetype != nil {
		for _,appId := range appIds {
			if application, appFound := c.applications[appId]; appFound {
				mimetype.AssociatedApplications = utils.AppendIfNotThere(mimetype.AssociatedApplications, appId)
				application.Mimetypes = utils.AppendIfNotThere(application.Mimetypes, mimeId)
			}
		}
	}
}

func (c* Collector) removeAssociations(mimeId string, appIds...string) {
	mimetype, mimetypeFound := c.mimetypes[mimeId]

	for _,appId := range appIds {
		if app, ok := c.applications[appId]; ok {
			app.Mimetypes = utils.Remove(app.Mimetypes, mimeId)
		}
		if mimetypeFound {
			mimetype.AssociatedApplications = utils.Remove(mimetype.AssociatedApplications, appId)
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
				} else if len(app.OnlyShowIn) > 0 &&
					      !utils.ElementsInCommon(xdg.CurrentDesktop, app.OnlyShowIn){
					delete(c.applications, app.Id)
				} else if len(app.NotShowIn) > 0 &&
					      utils.ElementsInCommon(xdg.CurrentDesktop, app.NotShowIn){
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

	iniFile, err := ini.ReadIniFile(path)
	if err != nil {
		if !os.IsNotExist(err) {
			fmt.Println(err)
		}
		mimeappsList.addedAssociations = make(map[string][]string)
		mimeappsList.removedAssociations = make(map[string][]string)
		mimeappsList.defaultApplications = make(map[string][]string)
	} else {
		for _, iniGroup := range iniFile.Groups {
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
			for _, entry := range iniGroup.Entries {
				dest[entry.Key] = utils.Split(entry.Value, ";")
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
		if mimetype := c.getMimetype(mimetypeId); mimetype != nil {
			for _,appId := range appIds {
				mimetype.DefaultApplications = append(mimetype.DefaultApplications, appId)
			}
		}
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


