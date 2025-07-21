package applications

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/server/icons"
	"github.com/surlykke/refude/server/lib/entity"
	"github.com/surlykke/refude/server/lib/icon"
	"github.com/surlykke/refude/server/lib/log"
	"github.com/surlykke/refude/server/lib/mediatype"
	"github.com/surlykke/refude/server/lib/slice"
	"github.com/surlykke/refude/server/lib/xdg"
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

		var id = strings.Replace(filePath[len(applicationsDir)+1:], "/", "-", -1)
		app, err := readDesktopFile(filePath, stripDesktopSuffix(id))
		if err != nil {
			log.Warn("Error processing ", filePath, ":\n\t", err)
			return nil
		}

		if app.Hidden ||
			(len(app.OnlyShowIn) > 0 && !slice.ElementsInCommon(xdg.CurrentDesktop, app.OnlyShowIn)) ||
			(len(app.NotShowIn) > 0 && slice.ElementsInCommon(xdg.CurrentDesktop, app.NotShowIn)) {
			return nil
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
				for _, appId := range slice.Split(appIds, ";") {
					appId = stripDesktopSuffix(appId)
					if app, ok := apps[appId]; ok {
						app.Mimetypes = slice.AppendIfNotThere(app.Mimetypes, mimetypeId)
					}
				}
			}
		}

		if removedAssociations := iniFile.FindGroup("Removed Associations"); removedAssociations != nil {
			for mimetypeId, appIds := range removedAssociations.Entries {
				for _, appId := range slice.Split(appIds, ";") {
					appId = stripDesktopSuffix(appId)
					if app, ok := apps[appId]; ok {
						app.Mimetypes = slice.Remove(app.Mimetypes, mimetypeId)
					}
				}
				if defaultAppIds, ok := defaultApps[mimetypeId]; ok {
					for _, appId := range slice.Split(appIds, ";") {
						appId = stripDesktopSuffix(appId)
						defaultAppIds = slice.Remove(defaultAppIds, appId)
					}
					defaultApps[mimetypeId] = defaultAppIds
				}
			}
		}

		if defaultApplications := iniFile.FindGroup("Default Applications"); defaultApplications != nil {
			for mimetypeId, defaultAppIds := range defaultApplications.Entries {
				var oldDefaultAppIds = defaultApps[mimetypeId]
				var newDefaultAppIds = make([]string, 0, len(defaultAppIds)+len(oldDefaultAppIds))
				for _, appId := range slice.Split(defaultAppIds, ";") {
					appId = stripDesktopSuffix(appId)
					newDefaultAppIds = slice.AppendIfNotThere(newDefaultAppIds, appId)
				}
				for _, appId := range oldDefaultAppIds {
					appId = stripDesktopSuffix(appId)
					newDefaultAppIds = slice.AppendIfNotThere(newDefaultAppIds, appId)
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

		var keywords = slice.Split(group.Entries["Keywords"], ";")
		var da = DesktopApplication{
			Base:      *entity.MakeBase(title, group.Entries["Comment"], icon.Name(iconName), mediatype.Application, keywords...),
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
		da.OnlyShowIn = slice.Split(group.Entries["OnlyShowIn"], ";")
		da.NotShowIn = slice.Split(group.Entries["NotShowIn"], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"] == "true"
		da.TryExec = group.Entries["TryExec"]
		da.Exec = group.Entries["Exec"]
		da.WorkingDir = group.Entries["Path"]
		da.Terminal = group.Entries["Terminal"] == "true"
		da.Categories = slice.Split(group.Entries["Categories"], ";")
		da.Implements = slice.Split(group.Entries["Implements"], ";")
		da.StartupNotify = group.Entries["StartupNotify"] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"]
		da.Url = group.Entries["URL"]
		da.Mimetypes = slice.Split(group.Entries["MimeType"], ";")
		da.DesktopFile = filePath
		da.AddAction("", "Open", "")
		da.DesktopActions = []DesktopAction{}
		var actionNames = slice.Split(group.Entries["Actions"], ";")

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Warn(da.DesktopId, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !slice.Contains(actionNames, currentAction) {
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

func stripDesktopSuffix(fileName string) string {
	return desktopSuffix.ReplaceAllString(fileName, "")
}
