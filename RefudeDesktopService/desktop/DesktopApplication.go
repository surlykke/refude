// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package desktop

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"golang.org/x/text/language"
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
	Mimetypes       []string
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
			reportError(fmt.Sprintf("Trying to run %s in terminal, but env variable TERMINAL not set", exec))
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
