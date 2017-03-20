package main

import (
	"errors"
	"github.com/surlykke/RefudeServices/common"
	"net/http"
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


func (app *DesktopApplication) Data(r *http.Request) (int, string, []byte) {
	if r.Method == "GET" {
		return common.GetJsonData(app)
	} else {
		return http.StatusMethodNotAllowed, "", nil
	}
}

func readDesktopFile(path string) (*DesktopApplication, common.StringSet, error) {
	iniGroups, err := common.ReadIniFile(path)

	if err != nil {
		return nil, nil, err
	}

	if iniGroups[0].Name != "Desktop Entry" {
		return nil, nil, errors.New("Desktop file should start with a 'Desktop Entry' group")
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
	app.Mimetypes = make(common.StringSet)
	app.Categories = common.ToSet(common.Split(desktopEntry["Categories"], ";"))
	app.Implements = common.ToSet(common.Split(desktopEntry["Implements"], ";"))
	app.Keywords = common.ToSet(common.Split(desktopEntry["Keywords"], ";"))
	app.Actions = make(map[string]Action, 0)

	return &app, common.ToSet(common.Split(desktopEntry["MimeType"], ";")), nil
}
