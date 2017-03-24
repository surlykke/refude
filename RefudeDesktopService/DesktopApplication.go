/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"errors"
	"github.com/surlykke/RefudeServices/common"
	"net/http"
	"fmt"
	"os/exec"
	"io/ioutil"
	"regexp"
	"strings"
	"github.com/surlykke/RefudeServices/service"
)

type DesktopApplication struct {
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	IconName        string `json:",omitempty"`
	IconUrl         string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      common.StringList
	NotShowIn       common.StringList
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       common.StringList
	Categories      common.StringList
	Implements      common.StringList
	Keywords        common.StringList
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]Action
	Id              string
}

type Action struct {
	Comment  string
	Name     string
	Exec     string
	IconName string
	IconUrl  string
}

func (app *DesktopApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, app)
	} else if r.Method == "POST" {
		actionId := "_default"
		actionv, ok := r.URL.Query()["action"]
		if ok && len(actionv) > 0{
			actionId = actionv[0]
		}

		if action,ok := app.Actions[actionId]; !ok {
			w.WriteHeader(http.StatusNotAcceptable)
		} else {
			cmd := regexp.MustCompile("%[uUfF]").ReplaceAllString(action.Exec, "")

			if err:= runCmd(cmd); err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusAccepted)
			}
		}
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func runCmd(app string) error {
	cmd := exec.Command("sh", "-c", "("+app+">/dev/null 2>/dev/null &)")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	ioutil.ReadAll(stderr)

	if err := cmd.Wait(); err != nil {
		return err
	}

	return nil
}



func readDesktopFile(path string) (*DesktopApplication, common.StringList, error) {
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
	app.Path = desktopEntry["Path"]
	app.StartupWmClass = desktopEntry["StartupWMClass"]
	app.Url = desktopEntry["URL"]

	if strings.HasPrefix(desktopEntry["Icon"], "/") {
		resourcePath := "/icon" + desktopEntry["Icon"]
		app.IconUrl = ".." + resourcePath
		service.Map(resourcePath, Icon{"/icon"})
	} else {
		app.IconName = desktopEntry["Icon"]
	}

	// FIXME use localized values
	app.Name = desktopEntry["Name"]
	app.GenericName = desktopEntry["GenericName"]
	app.Comment = desktopEntry["Comment"]

	app.OnlyShowIn = common.Split(desktopEntry["OnlyShowIn"], ";")
	app.NotShowIn = common.Split(desktopEntry["NotShowIn"], ";")
	app.Mimetypes = make(common.StringList, 0)
	app.Categories = common.Split(desktopEntry["Categories"], ";")
	app.Implements = common.Split(desktopEntry["Implements"], ";")
	app.Keywords = common.Split(desktopEntry["Keywords"], ";")
	app.Actions = make(map[string]Action, 0)
	app.Actions["_default"] = Action{
		Name: app.Name,
		Comment: app.Comment,
		IconName: app.IconName,
		IconUrl: app.IconUrl,
		Exec: desktopEntry["Exec"],
	}
	actionNames := common.Split(desktopEntry["Actions"], ";")
	for i := 1; i < len(iniGroups); i++ {
		if iniGroups[i].Name[0:15] != "Desktop Action " {
			continue
		} else if actionName := iniGroups[i].Name[15:]; !common.Find(actionNames, actionName) {
			fmt.Println("Unknown action", iniGroups[i].Name, " - ignoring")
			continue
		} else {
			action := Action{
				Name: iniGroups[i].Entry["Name"],
				Comment: app.Name,

				Exec: iniGroups[i].Entry["Exec"],
			}
			if strings.HasPrefix(iniGroups[i].Entry["Icon"], "/") {
				resourcePath := "/icon" + iniGroups[i].Entry["Icon"]
				action.IconUrl = ".." + resourcePath
				service.Map(resourcePath, Icon{"/icon"})
			} else {
				action.IconName = iniGroups[i].Entry["Icon"]
			}

			if action.IconUrl == "" && action.IconName == "" {
				action.IconUrl, action.IconName = app.IconUrl, app.IconName
			}

			app.Actions[actionName] = action
		}
	}

	return &app, common.Split(desktopEntry["MimeType"], ";"), nil
}
