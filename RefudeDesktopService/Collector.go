package main

import (
	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"path/filepath"
	"os"
	"strings"
	"github.com/surlykke/RefudeServices/lib/ini"
	"fmt"
)

type collection struct {
	applications map[string]*DesktopApplication
	mimetypes map[string]*Mimetype
}

func Collect() collection {
	c := collection{make(map[string]*DesktopApplication), make(map[string]*Mimetype)}
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

	return c
}

func (c *collection) getMimetype(id string) *Mimetype {
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


func (c*collection) addAssociations(mimeId string, appIds...string) {
	if mimetype := c.getMimetype(mimeId); mimetype != nil {
		for _,appId := range appIds {
			if application, appFound := c.applications[appId]; appFound {
				mimetype.AssociatedApplications = utils.AppendIfNotThere(mimetype.AssociatedApplications, appId)
				application.Mimetypes = utils.AppendIfNotThere(application.Mimetypes, mimeId)
			}
		}
	}
}

func (c*collection) removeAssociations(mimeId string, appIds...string) {
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

func (c *collection) collectApplications(appdir string) {
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

func (c *collection) readMimeappsList(path string) {
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
