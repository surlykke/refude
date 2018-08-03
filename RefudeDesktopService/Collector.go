// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"errors"
	"golang.org/x/sys/unix"
	"fmt"
	"github.com/surlykke/RefudeServices/lib"
)

type collection struct {
	mimetypes    map[string]*Mimetype           // Maps from mimetypeid to mimetype
	applications map[string]*DesktopApplication // Maps from applicationid to application
	defaultApps  map[string][]string            // Maps from mimetypeid to list of app ids
}

func CollectAndWatch(collected chan collection) {
	fd, err := unix.InotifyInit()
	defer unix.Close(fd)

	if err != nil {
		panic(err)
	}
	for _, dataDir := range append(lib.DataDirs, lib.DataHome) {
		appDir := dataDir + "/applications"
		fmt.Println("Watching: " + appDir)
		if _, err := unix.InotifyAddWatch(fd, appDir, unix.IN_CREATE|unix.IN_MODIFY|unix.IN_DELETE); err != nil {
			panic(err)
		}
	}

	if _, err := unix.InotifyAddWatch(fd, lib.ConfigHome+"/mimeapps.list", unix.IN_CLOSE_WRITE); err != nil {
		panic(err)
	}

	Collect(collected)
	dummy := make([]byte, 100)
	for {
		if _, err := unix.Read(fd, dummy); err != nil {
			panic(err)
		}
		fmt.Println("Something happened...")
		Collect(collected)
	}
}



func Collect(collected chan collection) {
	var c collection
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)
	c.defaultApps = make(map[string][]string)

	for _, dir := range lib.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(lib.DataHome + "/applications")

	for _, dir := range append(lib.ConfigDirs, lib.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for appId, app := range c.applications {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := c.mimetypes[mimetypeId]; ok {
				lib.AppendIfNotThere(mimetype.AssociatedApplications, appId)
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
	fmt.Println("Sending", len(c.mimetypes), len(c.applications))
	collected <- c
}

func (c *collection) removeApp(app *DesktopApplication) {
	delete(c.applications, app.Id)
	for _,mimetypeId := range app.Mimetypes {
		if mimetype,ok := c.mimetypes[mimetypeId]; ok {
			mimetype.AssociatedApplications = lib.Remove(mimetype.AssociatedApplications, app.Id)
		}
	}
}

func (c *collection) collectApplications(appdir string) {
	filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
		if !(info.IsDir() || !strings.HasSuffix(path, ".desktop")) {
			app, err := readDesktopFile(path)
			if err == nil {
				app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
				app.Self = lib.Standardizef("/applications/%s", app.Id)
				if oldApp, ok := c.applications[app.Id]; ok {
					c.removeApp(oldApp)
				}
				if !(app.Hidden ||
					(len(app.OnlyShowIn) > 0 && !lib.ElementsInCommon(lib.CurrentDesktop, app.OnlyShowIn)) ||
					(len(app.NotShowIn) > 0 && lib.ElementsInCommon(lib.CurrentDesktop, app.NotShowIn))) {
					delete(c.applications, app.Id)
					c.applications[app.Id] = app
					for _, mimetypeId := range app.Mimetypes {
						if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
							mimetype.AssociatedApplications = lib.AppendIfNotThere(mimetype.AssociatedApplications, app.Id)
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
	if iniFile, err := lib.ReadIniFile(path); err == nil {
		if addedAssociations := iniFile.FindGroup("Added Associations"); addedAssociations != nil {
			for mimetypeId, appIds := range addedAssociations.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					for _, appId := range lib.Split(appIds, ";") {
						if app, ok := c.applications[appId]; ok {
							app.Mimetypes = lib.AppendIfNotThere(app.Mimetypes, mimetypeId)
							mimetype.AssociatedApplications = lib.AppendIfNotThere(mimetype.AssociatedApplications, appId)
						}
					}
				}

			}
		}

		if removedAssociations := iniFile.FindGroup("Removed Associations"); removedAssociations != nil {
			for mimetypeId, appIds := range removedAssociations.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					for _, appId := range lib.Split(appIds, ";") {
						if app, ok := c.applications[appId]; ok {
							app.Mimetypes = lib.Remove(app.Mimetypes, mimetypeId)
							mimetype.AssociatedApplications = lib.Remove(mimetype.AssociatedApplications, appId)
						}
					}
				}

			}
		}

		if defaultApplications := iniFile.FindGroup("Default Applications"); defaultApplications != nil {
			for mimetypeId, appIds := range defaultApplications.Entries {
				if mimetype := c.getOrAdd(mimetypeId); mimetype != nil {
					var apps = lib.Split(appIds, ";")
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


func readDesktopFile(path string) (*DesktopApplication, error) {
	if iniFile, err := lib.ReadIniFile(path); err != nil {
		return nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, errors.New("File must start with '[Desktop Entry]'")
	} else {
		var da = DesktopApplication{}
		da.Mt = DesktopApplicationMediaType
		da.Actions = make(map[string]*Action)
		var actionNames = []string{}
		group := iniFile[0]

		if da.Type = group.Entries["Type"]; da.Type == "" {
			return nil, errors.New("Desktop file invalid, no 'Type' given")
		}
		da.Version = group.Entries["Version"]
		if da.Name = group.Entries["Name"]; da.Name == "" {
			return nil, errors.New("Desktop file invalid, no 'Name' given")
		}

		da.GenericName = group.Entries["GenericName"]
		da.NoDisplay = group.Entries["NoDisplay"] == "true"
		da.Comment = group.Entries["Comment"]
		icon := group.Entries["Icon"]
		if strings.HasPrefix(icon, "/") {
			if iconName, err := lib.CopyIconToSessionIconDir(icon); err != nil {
				log.Printf("Problem with iconpath %s in %s: %s", icon, da.Id, err.Error())
			} else {
				da.IconName = iconName
			}
		} else {
			da.IconName = icon
		}
		da.Hidden = group.Entries["Hidden"] == "true"
		da.OnlyShowIn = lib.Split(group.Entries["OnlyShowIn"], ";")
		da.NotShowIn = lib.Split(group.Entries["NotShowIn"], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"] == "true"
		da.TryExec = group.Entries["TryExec"]
		da.Exec = group.Entries["Exec"]
		da.Path = group.Entries["Path"]
		da.Terminal = group.Entries["Terminal"] == "true"
		actionNames = lib.Split(group.Entries["Actions"], ";")
		da.Mimetypes = lib.Split(group.Entries["MimeType"], ";")
		da.Categories = lib.Split(group.Entries["Categories"], ";")
		da.Implements = lib.Split(group.Entries["Implements"], ";")
		// FIXMEda.Keywords[tag] = utils.Split(group[""], ";")
		da.StartupNotify = group.Entries["StartupNotify"] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"]
		da.Url = group.Entries["URL"]

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Print(path, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !lib.Contains(actionNames, currentAction) {
				log.Print(path, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				var action Action
				if action.Name = actionGroup.Entries["Name"]; action.Name == "" {
					return nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				icon = actionGroup.Entries["Icon"]
				if strings.HasPrefix(icon, "/") {
					if iconName, err := lib.CopyIconToSessionIconDir(icon); err != nil {
						log.Printf("Problem with iconpath %s in %s: %s", icon, da.Id, err.Error())
					} else {
						action.IconName = iconName
					}
				} else {
					action.IconName = icon
				}
				action.Exec = actionGroup.Entries["Exec"]
				da.Actions[currentAction] = &action
			}
		}

		for _, action := range da.Actions {
			if action.IconName == "" {
				action.IconName = da.IconName
			}
		}

		return &da, nil
	}
}

