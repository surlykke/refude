// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/parser"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/text/language"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const DesktopApplicationMediaType resource.MediaType = "application/vnd.org.refude.desktopapplication+json"

type DesktopApplication struct {
	resource.AbstractResource
	Type            string
	Version         string `json:",omitempty"`
	Name            string
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Comment         string `json:",omitempty"`
	IconName        string `json:",omitempty"`
	Hidden          bool
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string `json:",omitempty"`
	Path            string `json:",omitempty"`
	Terminal        bool
	Categories      []string
	Implements      []string
	Keywords        []string
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	DesktopActions  map[string]*DesktopAction
	Id              string
	languages       language.Matcher
}

type DesktopAction struct {
	Name     string
	Exec     string
	IconName string
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {
	fmt.Println("In post")
	var actionName = requests.GetSingleQueryParameter(r, "action", "")
	var args = r.URL.Query()["arg"]
	var exec string
	if actionName == "" {
		exec = da.Exec
	} else if action, ok := da.DesktopActions[actionName]; !ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		exec = action.Exec
	}

	var onlySingleArg = !(strings.Contains(exec, "%F") || strings.Contains(exec, "%U"))
	if onlySingleArg && len(args) > 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		launchWithArgs(da.Exec, args, da.Terminal)
		w.WriteHeader(http.StatusAccepted)
	}
}

func launch(exec string, inTerminal bool) {
	launchWithArgs(exec, []string{}, inTerminal)
}

func launchWithArgs(exec string, args []string, inTerminal bool) {
	var argv []string
	var argsReg = regexp.MustCompile("%[uUfF]");
	if inTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			log.Println(fmt.Sprintf("Trying to run %s in terminal, but env variable TERMINAL not set", exec))
			return
		}
		var arglist = []string{}
		for _, arg := range args {
			arglist = append(arglist, "'"+strings.Replace(arg, "'", "'\\''", -1)+"'")
		}
		var argListS = strings.Join(arglist, " ");
		var cmd = argsReg.ReplaceAllString(exec, argListS)
		fmt.Println("Run in terminal with cmd:", cmd)
		argv = []string{terminal, "-e", cmd}
	} else {
		var fields = strings.Fields(exec)
		for _, field := range fields {
			if argsReg.MatchString(field) {
				argv = append(argv, args...)
			} else {
				argv = append(argv, field)
			}
		}
	}

	xdg.RunCmd(argv)
}

type DesktopApplicationCollection map[string]*DesktopApplication

func (dac DesktopApplicationCollection) GetResource(r *http.Request) (interface{}, error) {
	var path = r.URL.Path
	if path == "/applications" {
		var allApps = make([]*DesktopApplication, 0, len(dac))
		for _, app := range dac {
			allApps = append(allApps, app)
		}

		if params, err := requests.GetSingleParams(r, "q"); err != nil {
			return nil, err
		} else if matcher, err :=  parser.Parse(params["q"]); err != nil {
			return nil, err
		} else {
			var tmp = make([]*DesktopApplication, 0, len(allApps))
			for _, da := range allApps{
				if matcher(da) {
					tmp = append(tmp, da)
				}
			}
			allApps = tmp
		}

		return allApps, nil
	} else if strings.HasPrefix(path, "/application/") {
		return dac[path[len("/application/"):]], nil
	} else {
		return nil, nil
	}

}
