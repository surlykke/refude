package main

import (
	"errors"
	"github.com/surlykke/RefudeServices/common"
	"github.com/surlykke/RefudeServices/resources"
)

type DesktopApplication struct {
	resources.FallbackHandler
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	Icon            string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      common.StringSet
	NotShowIn       common.StringSet
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       common.StringSet
	Categories      common.StringSet
	Implements      common.StringSet
	Keywords        common.StringSet
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]Action
	Id              string
}

type Action struct {
	comment string
	name    string
	exec    string
	icon    string
}





func readDesktopFile(path string) (DesktopApplication, error) {
	iniGroups, err := common.ReadIniFile(path)

	if err != nil {
		return DesktopApplication{}, err
	}

	if iniGroups[0].Name != "Desktop Entry" {
		return DesktopApplication{}, errors.New("Desktop file should start with a 'Desktop Entry' group")
	}

	app := DesktopApplication{}
	desktopEntry := iniGroups[0].Entry
	app.Type = desktopEntry["Type"]
	app.Version = desktopEntry["Version"]
	app.Exec = desktopEntry["Exec"]
	app.Path = desktopEntry["Path"]
	app.StartupWmClass = desktopEntry["StartupWMClass"]
	app.Url = desktopEntry["URL"]
	app.Icon = desktopEntry["Icon"]

	// FIXME use localized values
	app.Name = desktopEntry["Name"]
	app.GenericName = desktopEntry["GenericName"]
	app.Comment = desktopEntry["Comment"]

	app.OnlyShowIn = common.ToSet(common.Split(desktopEntry["OnlyShowIn"], ";"))
	app.NotShowIn = common.ToSet(common.Split(desktopEntry["NotShowIn"], ";"))
	app.Mimetypes = common.ToSet(common.Split(desktopEntry["MimeType"], ";"))
	app.Categories = common.ToSet(common.Split(desktopEntry["Categories"], ";"))
	app.Implements = common.ToSet(common.Split(desktopEntry["Implements"], ";"))
	app.Keywords = common.ToSet(common.Split(desktopEntry["Keywords"], ";"))
	app.Actions = make(map[string]Action, 0)
	return app, nil
}
