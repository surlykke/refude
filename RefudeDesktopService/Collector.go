package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type association struct {
	mimetype    string
	application string
}

type collection struct {
	mimetypes    map[string]*Mimetype           // Maps from mimetypeid to mimetype
	applications map[string]*DesktopApplication // Maps from applicationid to application
	defaultApps  map[string][]string            // Maps from mimetypeid to list of app ids
}

func Collect() collection {
	var c collection
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)
	c.defaultApps = make(map[string][]string)

	for _, dir := range xdg.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome + "/applications")

	for _, dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for appId, app := range c.applications {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := c.mimetypes[mimetypeId]; ok {
				utils.AppendIfNotThere(mimetype.AssociatedApplications, appId)
			}
		}
	}

	for mimetypeId, appIds := range c.defaultApps {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if _, ok := c.applications[appId]; ok {
					mimetype.DefaultApplication = appId
					break
				}
			}
		}
	}

	return c
}

func (c *collection) removeApp(app *DesktopApplication) {
	delete(c.applications, app.Id)
	for _,mimetypeId := range app.Mimetypes {
		if mimetype,ok := c.mimetypes[mimetypeId]; ok {
			mimetype.AssociatedApplications = utils.Remove(mimetype.AssociatedApplications, app.Id)
		}
	}
}

func (c *collection) collectApplications(appdir string) {
	filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
		if !(info.IsDir() || !strings.HasSuffix(path, ".desktop")) {
			app, err := readDesktopFile(path)
			if err == nil {
				app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
				app.Self = "desktop-service:/applications/" + app.Id
				if oldApp, ok := c.applications[app.Id]; ok {
					c.removeApp(oldApp)
				}
				if !(app.Hidden ||
					(len(app.OnlyShowIn) > 0 && !utils.ElementsInCommon(xdg.CurrentDesktop, app.OnlyShowIn)) ||
					(len(app.NotShowIn) > 0 && utils.ElementsInCommon(xdg.CurrentDesktop, app.NotShowIn))) {
					delete(c.applications, app.Id)
					c.applications[app.Id] = app
					for _, mimetypeId := range app.Mimetypes {
						if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
							mimetype.AssociatedApplications = utils.AppendIfNotThere(mimetype.AssociatedApplications, app.Id)
						}
					}
				}
			} else {
				log.Println("Error processing ", path, ":\n\t", err)
			}
		}
		return nil
	})

}

type mimeAppsListReadState int

func (c *collection) getOrAdd(mimetypeId string) *Mimetype {
	if mimetype, ok := c.mimetypes[mimetypeId]; ok {
		return mimetype
	} else if mimetype, err := NewMimetype(mimetypeId); err == nil {
		c.mimetypes[mimetypeId] = mimetype
		return mimetype
	} else {
		log.Println(mimetypeId, "not legal")
		return nil
	}
}

func (c *collection) readMimeappsList(path string) {
	if iniFile, err := ini.ReadIniFile(path); err == nil {
		if addedAssociations := iniFile.FindGroup("Added Associations"); addedAssociations != nil {
			for mimetypeId, appIds := range addedAssociations.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					for _, appId := range utils.Split(appIds, ";") {
						if app, ok := c.applications[appId]; ok {
							app.Mimetypes = utils.AppendIfNotThere(app.Mimetypes, mimetypeId)
							mimetype.AssociatedApplications = utils.AppendIfNotThere(mimetype.AssociatedApplications, appId)
						}
					}
				}

			}
		}

		if removedAssociations := iniFile.FindGroup("Removed Associations"); removedAssociations != nil {
			for mimetypeId, appIds := range removedAssociations.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					for _, appId := range utils.Split(appIds, ";") {
						if app, ok := c.applications[appId]; ok {
							app.Mimetypes = utils.Remove(app.Mimetypes, mimetypeId)
							mimetype.AssociatedApplications = utils.Remove(mimetype.AssociatedApplications, appId)
						}
					}
				}

			}
		}

		if defaultApplications := iniFile.FindGroup("Default Applications"); defaultApplications != nil {
			for mimetypeId, appIds := range defaultApplications.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					var apps = utils.Split(appIds, ";")
					list, ok := c.defaultApps[mimetypeId]
					if !ok {
						list = make([]string, 0)
					}
					c.defaultApps[mimetypeId] = append(apps, list...)
				}
			}
		}
	}
}
