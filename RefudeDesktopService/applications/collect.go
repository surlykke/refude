// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/surlykke/RefudeServices/RefudeDesktopService/icons"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/text/language"
)

type collection struct {
	mimetypes    map[string]*Mimetype
	applications map[string]*DesktopApplication
	associations map[string][]string // Maps from mimetypeid to a list of app ids
	defaultApps  map[string][]string // Maps from mimetypeid to a list of app ids
}

func makeCollection() collection {
	return collection{
		mimetypes:    make(map[string]*Mimetype),
		applications: make(map[string]*DesktopApplication),
		associations: make(map[string][]string),
		defaultApps:  make(map[string][]string),
	}

}

func Collect() collection {
	var c = makeCollection()
	c.mimetypes = CollectMimeTypes()

	// Add aliases as mimetypes
	for _, mt := range c.mimetypes {
		for _, alias := range aliasTypes(mt) {
			if _, ok := c.mimetypes[alias.Id]; !ok {
				c.mimetypes[alias.Id] = alias
			}
			for _, appId := range c.associations[mt.Id] {
				c.associations[alias.Id] = slice.AppendIfNotThere(c.associations[alias.Id], appId)
			}
		}
	}

	for _, dir := range xdg.DataDirs {
		c.collectApplications(dir + "/applications")
		c.readMimeappsList(dir + "/applications/mimeapps.list")
	}

	c.collectApplications(xdg.DataHome + "/applications")

	for _, dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		c.readMimeappsList(dir + "/mimeapps.list")
	}

	for mimetypeId, appIds := range c.associations {
		if _, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if application, ok := c.applications[appId]; ok {
					application.Mimetypes = slice.AppendIfNotThere(application.Mimetypes, mimetypeId)
				}
			}
		}
	}

	for _, app := range c.applications {
		sort.Sort(slice.SortableStringSlice(app.Mimetypes))
	}

	// In case no default app is defined in a mimetypes.list somewhere
	// we take as default app any (randomly chosen) app that handles this mimetype
	for _, app := range c.applications {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := c.mimetypes[mimetypeId]; ok {
				mimetype.DefaultApp = app.Id
				break
			}
		}
	}

	for mimetypeId, appIds := range c.defaultApps {
		if mimetype, ok := c.mimetypes[mimetypeId]; ok {
			for _, appId := range appIds {
				if _, ok := c.applications[appId]; ok {
					mimetype.DefaultApp = appId
					mimetype.DefaultAppPath = appSelf(appId)
					break
				}
			}
		}
	}

	return c
}

func aliasTypes(mt *Mimetype) []*Mimetype {
	var result = make([]*Mimetype, 0, len(mt.Aliases))
	for _, id := range mt.Aliases {
		var copy = *mt
		copy.Id = id
		copy.Aliases = []string{}
		result = append(result, &copy)
	}

	return result
}

func (c *collection) removeAssociations(app *DesktopApplication) {
	for mimetypeId, appIds := range c.associations {
		c.associations[mimetypeId] = slice.Remove(appIds, app.Id)
	}
}

func CollectMimeTypes() map[string]*Mimetype {
	res := make(map[string]*Mimetype)

	for id, comment := range schemeHandlers {
		var mimetype, err = MakeMimetype(id)
		if err != nil {
			fmt.Println("Problem making mimetype", id)
		} else {
			mimetype.Comment = comment
			res[id] = mimetype
		}
	}

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

	for _, tmp := range xmlCollector.MimeTypes {
		if mimeType, err := MakeMimetype(tmp.Type); err != nil {
			fmt.Println(err)
		} else {
			var collectedLocales = make(map[string]bool)

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
				if xdg.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.ExpandedAcronym == "" {
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

			var tags = make([]language.Tag, len(collectedLocales))
			for locale := range collectedLocales {
				tags = append(tags, language.Make(locale))
			}

			res[mimeType.Id] = mimeType
		}
	}

	// Do a transitive closure on 'SubClassOf'
	for _, mt := range res {
		for i := 0; i < len(mt.SubClassOf); i++ {
			if ancestor, ok := res[mt.SubClassOf[i]]; ok {
				for _, id := range ancestor.SubClassOf {
					mt.SubClassOf = slice.AppendIfNotThere(mt.SubClassOf, id)
				}
			}
		}
	}

	return res
}

func (c *collection) collectApplications(appdir string) {
	var visitor = func(path string, info os.FileInfo, err error) error {
		if info.IsDir() || !strings.HasSuffix(path, ".desktop") {
			return nil
		}

		var id = strings.Replace(path[len(appdir)+1:], "/", "-", -1)
		app, mimetypes, err := readDesktopFile(path, id)
		if err != nil {
			log.Println("Error processing ", path, ":\n\t", err)
			return nil
		}

		if app.Hidden ||
			(len(app.OnlyShowIn) > 0 && !slice.ElementsInCommon(xdg.CurrentDesktop, app.OnlyShowIn)) ||
			(len(app.NotShowIn) > 0 && slice.ElementsInCommon(xdg.CurrentDesktop, app.NotShowIn)) {
			return nil
		}

		c.applications[app.Id] = app

		for _, mimetype := range mimetypes {
			c.associations[mimetype] = slice.AppendIfNotThere(c.associations[mimetype], app.Id)
		}
		return nil
	}

	if xdg.DirOrFileExists(appdir) {
		_ = filepath.Walk(appdir, visitor)
	}
}

func (c *collection) getOrAdd(mimetypeId string) *Mimetype {
	if mimetype, ok := c.mimetypes[mimetypeId]; ok {
		return mimetype
	} else if mimetype, err := MakeMimetype(mimetypeId); err == nil {
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

func readDesktopFile(path string, id string) (*DesktopApplication, []string, error) {
	if iniFile, err := xdg.ReadIniFile(path); err != nil {
		return nil, nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, nil, errors.New("File must start with '[Desktop Entry]'")
	} else {
		var da = DesktopApplication{Id: id}
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
			da.IconName = icons.AddFileIcon(icon)
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
		da.Mimetypes = []string{}

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Print(path, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !slice.Contains(actionNames, currentAction) {
				log.Print(path, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				var name = actionGroup.Entries["Name"]
				if name == "" {
					return nil, nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				var iconName = actionGroup.Entries["icon"]
				if strings.HasPrefix(iconName, "/") {
					iconName = icons.AddFileIcon(iconName)
				}
				da.DesktopActions[currentAction] = &DesktopAction{
					self:     actionPath(da.Id, currentAction),
					Name:     name,
					Exec:     actionGroup.Entries["Exec"],
					IconName: iconName,
				}
			}
		}
		mimetypes = slice.Split(group.Entries["MimeType"], ";")

		return &da, mimetypes, nil
	}
}
