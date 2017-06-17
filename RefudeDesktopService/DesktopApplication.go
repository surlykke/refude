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
	"net/http"
	"fmt"
	"os/exec"
	"io/ioutil"
	"regexp"
	"strings"
	"github.com/surlykke/RefudeServices/lib/service"
	"encoding/json"
	"io"
	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/utils"
	"github.com/surlykke/RefudeServices/lib/resource"
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
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       []string
	Categories      []string
	Implements      []string
	Keywords        []string
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]Action
	Id              string
	RelevanceHint   int
}

type Action struct {
	Comment  string
	Name     string
	Exec     string
	IconName string
	IconUrl  string
}

type DesktopPostPayload struct {
	ActionId string
	Arguments []string
}

func unmarshal(data io.Reader, dest interface{}) error {
	if extractedData, err := ioutil.ReadAll(data); err != nil {
		return err
	} else if len(extractedData) > 0 {
		return json.Unmarshal(extractedData, dest)
	} else {
		return nil
	}
}

func DesktopApplicationPOST(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
	app := this.Data.(*DesktopApplication)

	payload := DesktopPostPayload{}
	if err := unmarshal(r.Body, &payload); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusNotAcceptable);
		return
	}

	if len(payload.ActionId) == 0 {
		payload.ActionId = "_default"
	}

	fmt.Println("action: ", payload.ActionId, ", args")
	if action, ok := app.Actions[payload.ActionId]; !ok {
		w.WriteHeader(http.StatusNotAcceptable)
	} else {
		args := strings.Join(payload.Arguments, " ")
		cmd := regexp.MustCompile("%[uUfF]").ReplaceAllString(action.Exec, args)
		fmt.Println("Running cmd: " + cmd)
		if err := runCmd(cmd); err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
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

func readDesktopFile(path string) (*DesktopApplication, []string, error) {
	iniFile, err := ini.ReadIniFile(path)

	if err != nil {
		return nil, nil, err
	}

	if len(iniFile.Groups) < 1 {
		return nil, nil, errors.New("Empty desktopfile: " + path)
	}

	desktopEntry := iniFile.Groups[0]

	if desktopEntry.Name != "Desktop Entry" {
		return nil, nil, errors.New("Desktop file should start with a 'Desktop Entry' group: " + path)
	}

	app := DesktopApplication{}
	app.Type = desktopEntry.Value("Type")
	app.Version = desktopEntry.Value("Version")
	app.Path = desktopEntry.Value("Path")
	app.StartupWmClass = desktopEntry.Value("StartupWMClass")
	app.Url = desktopEntry.Value("URL")

	if strings.HasPrefix(desktopEntry.Value("Icon"), "/") {
		resourcePath := "/icon" + desktopEntry.Value("Icon")
		app.IconUrl = ".." + resourcePath
		service.Map(resourcePath, &resource.Resource{Data: "/icon", GET: IconGet})
	} else {
		app.IconName = desktopEntry.Value("Icon")
	}

	// FIXME use localized values
	app.Name = desktopEntry.Value("Name")
	app.GenericName = desktopEntry.Value("GenericName")
	app.Comment = desktopEntry.Value("Comment")

	app.OnlyShowIn = utils.Split(desktopEntry.Value("OnlyShowIn"), ";")

	app.NotShowIn = utils.Split(desktopEntry.Value("NotShowIn"), ";")
	app.Mimetypes = make([]string, 0)
	app.Categories = utils.Split(desktopEntry.Value("Categories"), ";")
	app.Implements = utils.Split(desktopEntry.Value("Implements"), ";")
	app.Keywords = utils.Split(desktopEntry.Value("Keywords"), ";")
	app.Actions = make(map[string]Action, 0)
	app.Actions["_default"] = Action{
		Name: app.Name,
		Comment: app.Comment,
		IconName: app.IconName,
		IconUrl: app.IconUrl,
		Exec: desktopEntry.Value("Exec"),
	}
	actionNames := utils.Split(desktopEntry.Value("Actions"), ";")
	for i := 1; i < len(iniFile.Groups); i++ {
		actionGroup := iniFile.Groups[i]
		if actionGroup.Name[0:15] != "Desktop Action " {
			continue
		} else if actionName := actionGroup.Name[15:]; !utils.Contains(actionNames, actionName) {
			fmt.Println("Unknown action", actionGroup.Name, " - ignoring")
			continue
		} else {
			action := Action{
				Name: actionGroup.Name,
				Comment: app.Name,

				Exec: actionGroup.Value("Exec"),
			}
			if strings.HasPrefix(actionGroup.Value("Icon"), "/") {
				resourcePath := "/icon" + actionGroup.Value("Icon")
				action.IconUrl = ".." + resourcePath
				service.Map(resourcePath, &resource.Resource{Data: "/icon", GET: IconGet})
			} else {
				action.IconName = actionGroup.Value("Icon")
			}

			if action.IconUrl == "" && action.IconName == "" {
				action.IconUrl, action.IconName = app.IconUrl, app.IconName
			}

			app.Actions[actionName] = action
		}
	}

	return &app, utils.Split(desktopEntry.Value("MimeType"), ";"), nil
}
