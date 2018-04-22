// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/service"
	"github.com/surlykke/RefudeServices/lib/utils"
	"golang.org/x/text/language"
	"os"
	"github.com/surlykke/RefudeServices/lib/xdg"
)


type DesktopApplication struct {
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	IconName        string `json:",omitempty"`
	IconPath        string `json:",omitempty"`
	IconUrl         string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Mimetypes       []string
	Categories      []string
	Implements      []string
	Keywords        []string
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	Actions         map[string]*Action
	Id              string
	RelevanceHint   int64
	languages       language.Matcher
	Self            string
	ResourceType    string
}

type Action struct {
	Name     string
	Exec     string
	IconName string
	IconPath string
	IconUrl  string
}

type IconPath string

func (ip IconPath) GET(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, string(ip))
}

func (da *DesktopApplication) GET(w http.ResponseWriter, r *http.Request) {
	resource.JsonGET(da, w)
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {
	actionId := resource.GetSingleQueryParameter(r, "action", "")
	var exec string
	if actionId != "" {
		if action, ok := da.Actions[actionId]; !ok {
			w.WriteHeader(http.StatusNotAcceptable)
			return
		} else {
			exec = action.Exec
		}
	} else {
		exec = da.Exec
	}
	var args = strings.Join(r.URL.Query()["arg"], " ")
	var argvAsString = regexp.MustCompile("%[uUfF]").ReplaceAllString(exec, args)
	if err := runCmd(da.Terminal, strings.Fields(argvAsString)); err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
		updatedApp := da
		updatedApp.RelevanceHint = time.Now().UnixNano() / 1000000
		service.Map(r.URL.Path, updatedApp)
	}

}

func runCmd(runInTerminal bool, argv []string) error {
	var cmd *exec.Cmd
	if runInTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			return errors.New("Trying to run " + strings.Join(argv, " ") + " in terminal, but env variable TERMINAL not set")
		}
		argv = append([]string{terminal, "-e"}, argv...)
	}
	cmd = exec.Command(argv[0], argv[1:]...)

	cmd.Dir = xdg.Home
	cmd.Stdout = nil
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	go cmd.Wait()

	return nil
}


func readDesktopFile(path string) (*DesktopApplication, error) {
	if iniFile, err := ini.ReadIniFile(path); err != nil {
		return nil, err
	} else if len(iniFile) == 0 || iniFile[0].Name != "Desktop Entry" {
		return nil, errors.New("File must start with '[Desktop Entry]'")
	} else {
		var da = DesktopApplication{ResourceType: "DesktopApplication"}
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
			da.IconPath = icon
			da.IconUrl = "../icons" + icon
		} else {
			da.IconName = icon
		}
		da.Hidden = group.Entries["Hidden"] == "true"
		da.OnlyShowIn = utils.Split(group.Entries["OnlyShowIn"], ";")
		da.NotShowIn = utils.Split(group.Entries["NotShowIn"], ";")
		da.DbusActivatable = group.Entries["DBusActivatable"] == "true"
		da.TryExec = group.Entries["TryExec"]
		da.Exec = group.Entries["Exec"]
		da.Path = group.Entries["Path"]
		da.Terminal = group.Entries["Terminal"] == "true"
		actionNames = utils.Split(group.Entries["Actions"], ";")
		da.Mimetypes = utils.Split(group.Entries["MimeType"], ";")
		da.Categories = utils.Split(group.Entries["Categories"], ";")
		da.Implements = utils.Split(group.Entries["Implements"], ";")
		// FIXMEda.Keywords[tag] = utils.Split(group[""], ";")
		da.StartupNotify = group.Entries["StartupNotify"] == "true"
		da.StartupWmClass = group.Entries["StartupWMClass"]
		da.Url = group.Entries["URL"]

		for _, actionGroup := range iniFile[1:] {
			if !strings.HasPrefix(actionGroup.Name, "Desktop Action ") {
				log.Print(path, ", ", "Unknown group type: ", actionGroup.Name, " - ignoring\n")
			} else if currentAction := actionGroup.Name[15:]; !utils.Contains(actionNames, currentAction) {
				log.Print(path, ", undeclared action: ", currentAction, " - ignoring\n")
			} else {
				var action Action
				if action.Name = actionGroup.Entries["Name"]; action.Name == ""{
					return nil, errors.New("Desktop file invalid, action " + actionGroup.Name + " has no default 'Name'")
				}
				icon = actionGroup.Entries["Icon"]
				if strings.HasPrefix(icon, "/") {
					action.IconPath = icon
					action.IconUrl = "../icons" + icon
				} else {
					action.IconName = icon
				}
				action.Exec = actionGroup.Entries["Exec"]
				da.Actions[currentAction] = &action
			}
		}


		for _, action := range da.Actions {
			if action.IconName == "" && action.IconPath == "" {
				action.IconName = da.IconName
				action.IconPath = da.IconPath
				action.IconUrl = da.IconUrl
			}
		}

		return &da, nil
	}
}

func transformLanguageTag(tag string) string {
	return strings.Replace(strings.Replace(tag, "_", "-", -1), "@", "-", -1)
}


