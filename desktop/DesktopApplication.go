package main

import (
	"errors"
)

type DesktopApplication struct {
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	Icon            string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      StringSet
	NotShowIn       StringSet
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       StringSet
	Categories      StringSet
	Implements      StringSet
	Keywords        StringSet
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
	iniGroups, err := ReadIniFile(path)

	if err != nil {
		return DesktopApplication{}, err
	}

	if iniGroups[0].name != "Desktop Entry" {
		return DesktopApplication{}, errors.New("Desktop file should start with a 'Desktop Entry' group")
	}

	app := DesktopApplication{}
	desktopEntry := iniGroups[0].entry
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

	app.OnlyShowIn = toSet(split(desktopEntry["OnlyShowIn"]))
	app.NotShowIn = toSet(split(desktopEntry["NotShowIn"]))
	app.Mimetypes = toSet(split(desktopEntry["MimeType"]))
	app.Categories = toSet(split(desktopEntry["Categories"]))
	app.Implements = toSet(split(desktopEntry["Implements"]))
	app.Keywords = toSet(split(desktopEntry["Keywords"]))
	app.Actions = make(map[string]Action, 0)
	return app, nil
}
