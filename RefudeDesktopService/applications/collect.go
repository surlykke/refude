// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"encoding/xml"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/text/language"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"errors"
	"fmt"
)

type collection struct {
	mimetypes    map[string]*Mimetype           // Maps from mimetypeid to mimetype
	applications map[string]*DesktopApplication // Maps from applicationid to application
	associations map[string][]string            // Maps from mimetypeid to a list of app ids
	defaultApps  map[string][]string            // Maps from mimetypeid to a list of app ids
}

func Collect() (map[string]*Mimetype, map[string]*DesktopApplication) {
	var c collection
	c.mimetypes = CollectMimeTypes()
	c.applications = make(map[string]*DesktopApplication)
	c.associations = make(map[string][]string) // Map a mimetypeid to a list of desktopapplication ids
	c.defaultApps = make(map[string][]string)  // Do

	for _, dir := range xdg.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome + "/applications")

	for _, dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for mimetypeId, appIds := range c.associations {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if application, ok := c.applications[appId]; ok {
					mimetype.LinkTo(resource.StandardizedPath("/application/" + appId), resource.Associated)
					application.LinkTo(resource.StandardizedPath("/mimetype/" + mimetypeId), resource.Associated)
				}
			}
		}
	}

	for mimetypeId, appIds := range c.defaultApps {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if _, ok := c.applications[appId]; ok {
					mimetype.LinkTo(resource.StandardizedPath("/application/" + appId), resource.DefaultApplication)
				}
			}
		}
	}


	fmt.Println("Collected ", len(c.mimetypes), "mimetypes and", len(c.applications), "applications")
	return c.mimetypes, c.applications
}

func (c *collection) removeAssociations(app *DesktopApplication) {
	for mimetypeId, appIds := range c.associations {
		c.associations[mimetypeId] = slice.Remove(appIds, app.Id)
	}
}

func CollectMimeTypes() map[string]*Mimetype {
	xmlCollector := struct {
		XMLName   xml.Name `xml:"mime-info"`
		MimeTypes []struct {
			Type    string `xml:"type,attr"`
			Comment []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"comment"`
			Acronym []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"acronym"`
			ExpandedAcronym []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"expanded-acronym"`
			Alias []struct {
				Type string `xml:"type,attr"`
			} `xml:"alias"`
			Glob []struct {
				Pattern string `xml:"pattern,attr"`
			} `xml:"glob"`
			SubClassOf []struct {
				Type string `xml:"type,attr"`
			} `xml:"sub-class-of"`
			Icon struct {
				Name string `xml:"name,attr"`
			} `xml:"icon"`
			GenericIcon struct {
				Name string `xml:"name,attr"`
			} `xml:"generic-icon"`
		} `xml:"mime-type"`
	}{}

	xmlInput, err := ioutil.ReadFile(freedesktopOrgXml)
	if err != nil {
		fmt.Println("Unable to open ", freedesktopOrgXml, ": ", err)
	}
	parseErr := xml.Unmarshal(xmlInput, &xmlCollector)
	if parseErr != nil {
		fmt.Println("Error parsing: ", parseErr)
	}

	res := make(map[string]*Mimetype)
	for _, tmp := range xmlCollector.MimeTypes {
		if mimeType, err := NewMimetype(tmp.Type); err != nil {
			fmt.Println(err)
		} else {
			var collectedLocales= make(map[string]bool)

			for _, tmpComment := range tmp.Comment {
				if xdg.LocaleMatch(tmpComment.Lang) || (tmpComment.Lang == "" && mimeType.Comment == "") {
					mimeType.Comment = tmpComment.Text
				}
			}

			for _, tmpAcronym := range tmp.Acronym {
				if xdg.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && mimeType.Acronym == "") {
					mimeType.Acronym = tmpAcronym.Text
				}
			}

			for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
				if xdg.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.Acronym == "" {
					mimeType.ExpandedAcronym = tmpExpandedAcronym.Text
				}
			}

			if tmp.Icon.Name != "" {
				mimeType.IconName = tmp.Icon.Name
			} else {
				mimeType.IconName = strings.Replace(mimeType.Id, "/", "-", -1)
			}

			for _, aliasStruct := range tmp.Alias {
				mimeType.Aliases = slice.AppendIfNotThere(mimeType.Aliases, aliasStruct.Type)
			}

			for _, tmpGlob := range tmp.Glob {
				mimeType.Globs = slice.AppendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
			}

			for _, tmpSubClassOf := range tmp.SubClassOf {
				mimeType.SubClassOf = slice.AppendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
			}

			if tmp.GenericIcon.Name != "" {
				mimeType.GenericIcon = tmp.GenericIcon.Name
			} else {
				slashPos := strings.Index(mimeType.Id, "/")
				mimeType.GenericIcon = mimeType.Id[:slashPos] + "-x-generic"
			}

			var tags= make([]language.Tag, len(collectedLocales))
			for locale, _ := range collectedLocales {
				tags = append(tags, language.Make(locale))
			}

			res[mimeType.Id] = mimeType
		}
	}

	return res
}

func (c *collection) collectApplications(appdir string) {
	if xdg.DirOrFileExists(appdir) {
		_ = filepath.Walk(appdir, func(path string, info os.FileInfo, err error) error {
			if strings.HasSuffix(path, ".desktop") && !info.IsDir() {
				app, mimetypes, err := readDesktopFile(path)
				if err == nil {
					app.Id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
					app.AbstractResource = resource.MakeAbstractResource(resource.Standardizef("/applications/%s", app.Id), DesktopApplicationMediaType)
					var exec = app.Exec
					var inTerminal = app.Terminal
					app.ResourceActions["default"] = resource.ResourceAction{
						Description: "Launch", IconName: app.IconName, Executer: func() { launch(exec, inTerminal) },
					}
					for id, action := range app.DesktopActions {
						var exec = action.Exec
						var inTerminal = app.Terminal
						app.ResourceActions[id] = resource.ResourceAction{
							Description: action.Name, IconName: action.IconName, Executer: func() { launch(exec, inTerminal) },
						}
					}

					c.applications[app.Id] = app

					for _, mimetype := range mimetypes {
						c.associations[mimetype] = slice.AppendIfNotThere(c.associations[mimetype], app.Id)
					}

				} else {
					log.Println("Error processing ", path, ":\n\t", err)
				}
			}
			return nil
		})
	}
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
	if iniFile, err := xdg.ReadIniFile(path); err == nil {
		if addedAssociations := iniFile.FindGroup("Added Associations"); addedAssociations != nil {
			for mimetypeId, appIds := range addedAssociations.Entries {
				for _, appId := range slice.Split(appIds, ";") {
					c.associations[mimetypeId] = slice.AppendIfNotThere(c.associations[mimetypeId], appId)
				}
			}
		}

		if removedAssociations := iniFile.FindGroup("Removed Associations"); removedAssociations != nil {
			for mimetypeId, appIds := range removedAssociations.Entries {
				for _, appId := range slice.Split(appIds, ";") {
					c.associations[mimetypeId] = slice.Remove(c.associations[mimetypeId], appId)
				}
			}
		}

		if defaultApplications := iniFile.FindGroup("Default Applications"); defaultApplications != nil {
			for mimetypeId, appIds := range defaultApplications.Entries {
				var tmp = c.defaultApps[mimetypeId]
				c.defaultApps[mimetypeId] = []string{}
				for _, appId := range append(slice.Split(appIds, ";"), tmp...) {
					c.defaultApps[mimetypeId] = slice.AppendIfNotThere(c.defaultApps[mimetypeId], appId)
				}
			}
		}
	}
}

func readDesktopFile(path string) (*DesktopApplication, []string, error) {
	if iniFile, err := xdg.ReadIniFile(path); err != nil {
		return nil, nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, nil, errors.New("File must start with '[Desktop Entry]'")
	} else {
		var da = DesktopApplication{}
		var mimetypes = []string{}
		da.DesktopActions = make(map[string]*DesktopAction)
		var actionNames = []string{}
		group := iniFile[0]

		if da.Type = group.Entries["Type"]; da.Type == "" {
			return nil, nil, errors.New("Desktop file invalid, no 'Type' given")
		}
		da.Version = group.Entries["Version"]
		if da.Name = group.Entries["Name"]; da.Name == "" {
			return nil, nil, errors.New("Desktop file invalid, no 'Name' given")
		}

		da.GenericName = group.Entries["GenericName"]
		da.NoDisplay = group.Entries["NoDisplay"] == "true"
		da.Comment = group.Entries["Comment"]
		icon := group.Entries["Icon"]
		if strings.HasPrefix(icon, "/") {
			if iconName, err := image.CopyIconToSessionIconDir(icon); err != nil {
				log.Printf("Problem with iconpath %s in %s: %s", icon, da.Id, err.Error())
			} else {
				da.IconName = iconName
			}
		} else {
			da.IconName = icon
		}
		da.Hidden = group.Entries["Hidden"] == "true"
		da.OnlyShowIn = slice.Split(group.Entries["OnlyShowIn"], ";")
		da.NotShowIn = slice.Split(group.Entries["NotShowIn"], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"] == "true"
		da.TryExec = group.Entries["TryExec"]
		da.Exec = group.Entries["Exec"]
		da.Path = group.Entries["Path"]
		da.Terminal = group.Entries["Terminal"] == "true"
		actionNames = slice.Split(group.Entries["Actions"], ";")
		da.Categories = slice.Split(group.Entries["Categories"], ";")
		da.Implements = slice.Split(group.Entries["Implements"], ";")
		// FIXMEda.Keywords[tag] = utils.Split(group[""], ";")
		da.StartupNotify = group.Entries["StartupNotify"] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"]
		da.Url = group.Entries["URL"]

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Print(path, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !slice.Contains(actionNames, currentAction) {
				log.Print(path, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				var action DesktopAction
				if action.Name = actionGroup.Entries["Name"]; action.Name == "" {
					return nil, nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				icon = actionGroup.Entries["Icon"]
				if strings.HasPrefix(icon, "/") {
					if iconName, err := image.CopyIconToSessionIconDir(icon); err != nil {
						log.Printf("Problem with iconpath %s in %s: %s", icon, da.Id, err.Error())
					} else {
						action.IconName = iconName
					}
				} else {
					action.IconName = icon
				}
				if action.IconName == "" {
					action.IconName = da.IconName
				}
				action.Exec = actionGroup.Entries["Exec"]
				da.DesktopActions[currentAction] = &action
			}
		}

		mimetypes = slice.Split(group.Entries["MimeType"], ";")

		return &da, mimetypes, nil
	}
}
