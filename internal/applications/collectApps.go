// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/icons"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/log"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/lib/xdg"
)

func collectApps() (map[string]*DesktopApplication, map[string][]string) {
	// ----
	// Keyed by apps DeskoptId, ie. 'firefox.desktop' not '/application/firefox.desktop'
	var apps = make(map[string]*DesktopApplication)
	// Mimetype id (ie. 'text/html' not '/mimetype/text/html') to list of DesktopIds
	var defaultApps = make(map[string][]string)

	for i := len(xdg.DataDirs) - 1; i >= 0; i-- {
		var dir = xdg.DataDirs[i]
		collectApplications(dir+"/applications", apps)
		readMimeappsList(dir+"/applications/mimeapps.list", apps, defaultApps)
	}

	collectApplications(xdg.DataHome+"/applications", apps)

	for _, dir := range append(xdg.ConfigDirs, xdg.ConfigHome) {
		readMimeappsList(dir+"/mimeapps.list", apps, defaultApps)
	}
	return apps, defaultApps
}

func collectApplications(applicationsDir string, apps map[string]*DesktopApplication) {
	var visitor = func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(filePath, ".desktop") {
			return nil
		}

		var id = strings.ReplaceAll(filePath[len(applicationsDir)+1:], "/", "-")
		app, err := readDesktopFile(filePath, trimAndStripDesktopSuffix(id))
		if err != nil {
			log.Warn("Error processing ", filePath, ":\n\t", err)
			return nil
		}

		if app.Hidden {
			return nil
		}

		if len(app.OnlyShowIn) > 0 && len(xdg.CurrentDesktop) > 0 {
			var match = false
			for _, osi := range app.OnlyShowIn {
				if slices.Contains(xdg.CurrentDesktop, osi) {
					match = true
					break
				}

			}
			if !match {
				return nil
			}
		}

		for _, d := range xdg.CurrentDesktop {
			if slices.Contains(app.NotShowIn, d) {
				return nil
			}
		}

		var executableName = app.Exec
		if lastSlash := strings.LastIndex(executableName, "/"); lastSlash > -1 {
			executableName = executableName[lastSlash:]
		}
		//app.Keywords = append(app.Keywords, executableName)

		apps[app.DesktopId] = app

		return nil
	}

	if xdg.DirOrFileExists(applicationsDir) {
		_ = filepath.Walk(applicationsDir, visitor)
	}
}

func readMimeappsList(path string, apps map[string]*DesktopApplication, defaultApps map[string][]string) {
	if iniFile, err := xdg.ReadIniFile(path); err == nil {
		if addedAssociations := iniFile.FindGroup("Added Associations"); addedAssociations != nil {
			for mimetypeId, appIds := range addedAssociations.Entries {
				for _, appId := range utils.Split(appIds, ";") {
					appId = trimAndStripDesktopSuffix(appId)
					if app, ok := apps[appId]; ok {
						app.Mimetypes = appendIfNotThere(app.Mimetypes, mimetypeId)
					}
				}
			}
		}

		if removedAssociations := iniFile.FindGroup("Removed Associations"); removedAssociations != nil {
			for mimetypeId, appIds := range removedAssociations.Entries {
				for _, appId := range utils.Split(appIds, ";") {
					appId = trimAndStripDesktopSuffix(appId)
					if app, ok := apps[appId]; ok {
						app.Mimetypes = remove(app.Mimetypes, mimetypeId)
					}
				}
				if defaultAppIds, ok := defaultApps[mimetypeId]; ok {
					for _, appId := range utils.Split(appIds, ";") {
						appId = trimAndStripDesktopSuffix(appId)
						defaultAppIds = remove(defaultAppIds, appId)
					}
					defaultApps[mimetypeId] = defaultAppIds
				}
			}
		}

		if defaultApplications := iniFile.FindGroup("Default Applications"); defaultApplications != nil {
			for mimetypeId, defaultAppIds := range defaultApplications.Entries {
				var oldDefaultAppIds = defaultApps[mimetypeId]
				var newDefaultAppIds = make([]string, 0, len(defaultAppIds)+len(oldDefaultAppIds))
				for _, appId := range utils.Split(defaultAppIds, ";") {
					appId = trimAndStripDesktopSuffix(appId)
					newDefaultAppIds = appendIfNotThere(newDefaultAppIds, appId)
				}
				for _, appId := range oldDefaultAppIds {
					appId = trimAndStripDesktopSuffix(appId)
					newDefaultAppIds = appendIfNotThere(newDefaultAppIds, appId)
				}
				defaultApps[mimetypeId] = newDefaultAppIds
			}
		}
	}

}

func readDesktopFile(filePath string, id string) (*DesktopApplication, error) {
	if iniFile, err := xdg.ReadIniFile(filePath); err != nil {
		return nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, errors.New("file must start with '[Desktop Entry]'")
	} else {
		group := iniFile[0]

		var title, iconName = group.Entries["Name"], group.Entries["Icon"]
		if strings.HasPrefix(string(iconName), "/") {
			icons.AddFileIcon(iconName)
		} else if strings.HasSuffix(iconName, ".png") || strings.HasSuffix(iconName, ".svg") || strings.HasSuffix(iconName, ".xpm") {
			iconName = iconName[:len(iconName)-4]
		}

		if title == "" {
			return nil, errors.New("desktop file invalid, no 'Name' given")
		}

		var keywords = utils.Split(group.Entries["Keywords"], ";")
		var da = DesktopApplication{
			Base:      *entity.MakeBase(title, group.Entries["Comment"], icon.Name(iconName), entity.Application, keywords...),
			DesktopId: id,
		}

		da.Comment = group.Entries["Comment"]
		if da.Type = group.Entries["Type"]; da.Type == "" {
			return nil, errors.New("desktop file invalid, no 'Type' given")
		}
		da.Version = group.Entries["Version"]
		da.GenericName = group.Entries["GenericName"]
		da.NoDisplay = group.Entries["NoDisplay"] == "true"
		da.Hidden = group.Entries["Hidden"] == "true"
		da.OnlyShowIn = utils.Split(group.Entries["OnlyShowIn"], ";")
		da.NotShowIn = utils.Split(group.Entries["NotShowIn"], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"] == "true"
		da.TryExec = group.Entries["TryExec"]
		da.Exec = group.Entries["Exec"]
		da.WorkingDir = group.Entries["Path"]
		da.Terminal = group.Entries["Terminal"] == "true"
		da.Categories = utils.Split(group.Entries["Categories"], ";")
		da.Implements = utils.Split(group.Entries["Implements"], ";")
		da.StartupNotify = group.Entries["StartupNotify"] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"]
		da.Url = group.Entries["URL"]
		da.Mimetypes = utils.Split(group.Entries["MimeType"], ";")
		da.DesktopFile = filePath
		da.AddAction("", "Open", "")
		da.DesktopActions = []DesktopAction{}
		var actionNames = utils.Split(group.Entries["Actions"], ";")

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Warn(da.DesktopId, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !slices.Contains(actionNames, currentAction) {
				log.Warn(da.DesktopId, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				var name = actionGroup.Entries["Name"]
				if name == "" {
					return nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				var iconUrl = icon.Name(actionGroup.Entries["Icon"])
				da.DesktopActions = append(da.DesktopActions, DesktopAction{
					id:   currentAction,
					Name: name,
					Exec: actionGroup.Entries["Exec"],
					Icon: iconUrl,
				})
				da.AddAction(currentAction, name, iconUrl)
			}
		}

		return &da, nil
	}

}

var desktopSuffix = regexp.MustCompile(`\.desktop$`)

func trimAndStripDesktopSuffix(fileName string) string {
	return strings.TrimSuffix(strings.TrimSpace(fileName), ".desktop")
}
