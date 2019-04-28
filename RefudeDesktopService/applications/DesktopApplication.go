// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/serialize"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

const DesktopApplicationMediaType resource.MediaType = "application/vnd.org.refude.desktopapplication+json"

type DesktopApplication struct {
	resource.GenericResource
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
	Mimetypes       []string
}

type DesktopAction struct {
	Name     string
	Exec     string
	IconName string
}

func (da *DesktopApplication) POST(w http.ResponseWriter, r *http.Request) {
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
	var argsReg = regexp.MustCompile("%[uUfF]")
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
		var argListS = strings.Join(arglist, " ")
		var cmd = argsReg.ReplaceAllString(exec, argListS)
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

func appSelf(appId string) resource.StandardizedPath {
	return resource.Standardizef("/applications/%s", appId)
}

func (da *DesktopApplication) WriteBytes(w io.Writer) {
	da.GenericResource.WriteBytes(w)
	serialize.String(w, da.Type)
	serialize.String(w, da.Version)
	serialize.String(w, da.Name)
	serialize.String(w, da.GenericName)
	serialize.Bool(w, da.NoDisplay)
	serialize.String(w, da.Comment)
	serialize.String(w, da.IconName)
	serialize.Bool(w, da.Hidden)
	serialize.StringSlice(w, da.OnlyShowIn)
	serialize.StringSlice(w, da.NotShowIn)
	serialize.Bool(w, da.DbusActivatable)
	serialize.String(w, da.TryExec)
	serialize.String(w, da.Exec)
	serialize.String(w, da.Path)
	serialize.Bool(w, da.Terminal)
	serialize.StringSlice(w, da.Categories)
	serialize.StringSlice(w, da.Implements)
	serialize.StringSlice(w, da.Keywords)
	serialize.Bool(w, da.StartupNotify)
	serialize.String(w, da.StartupWmClass)
	serialize.String(w, da.Url)
	for id, action := range da.DesktopActions {
		serialize.String(w, id)
		action.WriteBytes(w)
	}
	serialize.String(w, da.Id)
	serialize.StringSlice(w, da.Mimetypes)
}

func (action *DesktopAction) WriteBytes(w io.Writer) {
	serialize.String(w, action.Name)
	serialize.String(w, action.Exec)
	serialize.String(w, action.IconName)
}
