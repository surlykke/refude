package main

import (
	"os"
	"path/filepath"
	"refude-service/xdg"
	"strings"
	"fmt"
	"time"
)

type AppMap map[string]DesktopApplication
type MimeMap map[string]MimeType


func CollectFromDesktop() (MimeMap, AppMap)  {
	fmt.Println(time.Now(), "Into NewDesktopCollection")
	c := desktopCollection{}
	c.Mimes = CollectMimeTypes()
	fmt.Println(time.Now(), "Mimes collected")
	c.Apps = make(map[string]DesktopApplication)
	c.mimeId2associatedApps = make(map[string]*StringSet)
	c.defaultApps = make(map[string][]string)

	for _, dir := range xdg.DataDirs() {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome() + "/applications")

	for _,dir := range append(xdg.ConfigDirs(), xdg.ConfigHome()) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	c.postProcess()
	fmt.Println(time.Now(), "Update done")
	return c.Mimes, c.Apps
}

type desktopCollection struct {
	Apps  AppMap
	Mimes MimeMap

	mimeId2associatedApps   map[string]*StringSet
	defaultApps map[string][]string
}

func (c* desktopCollection) addAssociations(mimeId string, appIds...string) {
	appSet, ok	:= c.mimeId2associatedApps[mimeId]
	if !ok {
		tmp := make(StringSet)
		appSet = &tmp
	}
	appSet.addAll(appIds)
	c.mimeId2associatedApps[mimeId] = appSet
}

func (c* desktopCollection) removeAssociations(mimeId string, appIds...string) {
	appSet, ok := c.mimeId2associatedApps[mimeId]
	if ok {
		appSet.removeAll(appIds)
		c.mimeId2associatedApps[mimeId] = appSet
	}
}

func (c *desktopCollection) collectApplications(appdir string) {
	filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
		if !(info.IsDir() || !strings.HasSuffix(path, ".desktop")) {
			app, err := readDesktopFile(path)
			if err == nil {
				app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
				if (app.Hidden) {
					delete(c.Apps, app.Id)
				} else {
					for mimetypeId,_ := range app.Mimetypes {
						c.addAssociations(mimetypeId, app.Id)
					}
					app.Mimetypes = make(StringSet)
					c.Apps[app.Id] = app
				}
			}
		}
		return nil
	})

}

func (c *desktopCollection) readMimeappsList(path string) {
    mimeappsList := struct {
		defaultApplications map[string][]string
		addedAssociations   map[string][]string
		removedAssociations map[string][]string
	}{make(map[string][]string), make(map[string][]string), make(map[string][]string)}

	iniGroups, err := ReadIniFile(path)
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
			switch iniGroup.name {
			case "Default Applications":
				dest = mimeappsList.defaultApplications
			case "Added Associations":
				dest = mimeappsList.addedAssociations
			case "Removed Associations":
				dest = mimeappsList.removedAssociations
			default:
				continue
			}
			for k, v := range iniGroup.entry {
				dest[k] = split(v)
			}
		}
	}

	for mimeId, appIds := range mimeappsList.removedAssociations {
		c.removeAssociations(mimeId, appIds...)
	}

	for mimeId, appIds := range mimeappsList.addedAssociations {
		c.addAssociations(mimeId, appIds...)
	}

	for mimetype, appIds := range mimeappsList.defaultApplications {
		c.defaultApps[mimetype] = append(c.defaultApps[mimetype], appIds...)
	}
}

func (c *desktopCollection) postProcess() {
	for mimeId, appIds := range c.mimeId2associatedApps {
		for appId,_ := range *appIds {
			if app, ok := c.Apps[appId]; ok {
				app.Mimetypes.add(mimeId)
			}
			if mime, ok := c.Mimes[mimeId]; ok {
				mime.AssociatedApplications.add(appId)
			}
		}
	}

	for mimetypeId, appList := range c.defaultApps {
		if mimetype, ok := c.Mimes[mimetypeId]; ok {
			mimetype.DefaultApplications = removeDublets(appList)
			c.Mimes[mimetypeId] = mimetype
		}
	}
}


