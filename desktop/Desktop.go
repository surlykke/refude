package main

import (

	"os"
	"path/filepath"
	"strings"
	"fmt"
	"github.com/surlykke/RefudeServices/xdg"
	"github.com/surlykke/RefudeServices/common"
	"github.com/surlykke/RefudeServices/service"
	"reflect"
	"golang.org/x/sys/unix"
)



type Desktop struct {
	applications map[string]*DesktopApplication
	mimetypes map[string]*Mimetype}

func NewDesktop() Desktop {
	service.Map("/applications", make(common.StringSet))
	service.Map("/mimetypes", make(common.StringSet))
	return Desktop{make(map[string]*DesktopApplication), make(map[string]*Mimetype)}
}


func (d *Desktop) Run() {

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
	d.Update()
	dummy := make([]byte, 100)
	for {
		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Something happened...")
	}
}

func (d *Desktop) Update() {
	c := NewCollector()
	c.collect()


	for appId,_ := range d.applications {
		if _,ok := c.applications[appId]; ok {
			service.Unmap("/application/" + appId)
		}
	}

	appPaths := make(common.StringSet)
	for appId, newDesktopApplication := range c.applications {
		path := "/application/" + appId
		if oldDesktopApplication,ok := d.applications[appId]; !ok {
			service.Map(path, newDesktopApplication)
		} else {
			if !reflect.DeepEqual(oldDesktopApplication, newDesktopApplication) {
				service.Remap(path, newDesktopApplication)
			}
		}
		appPaths[path[1:]] = true
	}
	service.Remap("/applications", appPaths)

	for mimeId,_ := range d.mimetypes {
		if _,ok := c.mimetypes[mimeId]; !ok {
			service.Unmap("/mimetype/" + mimeId)
		}
	}

	mimePaths := make(common.StringSet)
	for mimeId, mimeType := range c.mimetypes {
		path := "/mimetype/" + mimeId
		if oldMimetype, ok := d.mimetypes[mimeId]; !ok {
			service.Map(path, mimeType)
		} else {
			if !reflect.DeepEqual(oldMimetype, mimeType) {
				service.Remap(path, mimeType)
			}
		}
		mimePaths[path[1:]] = true
	}
	service.Remap("/mimetypes", mimePaths)

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

	for _, dir := range xdg.DataDirs() {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome() + "/applications")

	for _,dir := range append(xdg.ConfigDirs(), xdg.ConfigHome()) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}
}

func appPath(id string) string {
	return "/application/" + id
}


func (c* Collector) addAssociations(mimeId string, appIds...string) {
	if mimetype, mimetypeFound := c.mimetypes[mimeId]; mimetypeFound {
		for _,appId := range appIds {
			if application, appFound := c.applications[appId]; appFound {
				mimetype.AssociatedApplications[appId] = true
				application.Mimetypes[mimeId] = true
			}
		}
	}
}

func (c* Collector) removeAssociations(mimeId string, appIds...string) {
	mimetype, mimetypeFound := c.mimetypes[mimeId]

	for _,appId := range appIds {
		if app, ok := c.applications[appId]; ok {
			delete(app.Mimetypes, mimeId)
		}
		if mimetypeFound {
			delete(mimetype.AssociatedApplications, appId)
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
					for mimetypeId,_ := range mimetypes {
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


