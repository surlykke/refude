// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"

	"github.com/surlykke/RefudeServices/lib/xdg"
)

type DesktopApplication struct {
	self            string
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
	DesktopActions  []DesktopAction
	Id              string
	Mimetypes       []string
}

func (d *DesktopApplication) ToStandardFormat() *respond.StandardFormat {
	var self = d.self
	var actions []respond.Action
	if len(d.DesktopActions) > 0 {
		for _, da := range d.DesktopActions {
			actions = append(actions, respond.Action{Title: da.Name, IconName: da.IconName, Path: d.self + "?actionid=" + da.id})
		}
	}
	return &respond.StandardFormat{
		Self:      self,
		OnPost:    "Launch",
		Actions:   actions,
		Type:      "application",
		Title:     d.Name,
		Comment:   d.Comment,
		IconName:  d.IconName,
		Data:      d,
		NoDisplay: d.NoDisplay,
	}
}

func (d *DesktopApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, d.ToStandardFormat())
	} else if r.Method == "POST" {

		var exec = d.Exec
		var actionId = requests.GetSingleQueryParameter(r, "actionid", "")
		if actionId != "" {
			if action, ok := d.action(actionId); ok {
				exec = action.Exec
			} else {
				respond.UnprocessableEntity(w, fmt.Errorf("Invalid actionid: %s", actionId))
				return
			}
		}

		respond.AcceptedAndThen(w, func() { run(exec, "", d.Terminal) })
	} else {
		respond.NotAllowed(w)
	}
}

func (d *DesktopApplication) Run(arg string) error {
	return run(d.Exec, arg, d.Terminal)
}

type DesktopAction struct {
	id       string
	Name     string
	Exec     string
	IconName string
}

func (da *DesktopAction) Run(arg string) error {
	return run(da.Exec, arg, false)
}

func (d *DesktopApplication) action(id string) (DesktopAction, bool) {
	for _, a := range d.DesktopActions {
		if a.id == id {
			return a, true
		}
	}

	return DesktopAction{}, false
}

func OpenFile(path string, mimetypeId string) error {
	var c = collectionStore.Load().(collection)
	if mt, ok := c.mimetypes[mimetypeId]; ok {
		if mt.DefaultApp != "" {
			if app, ok := c.applications[mt.DefaultApp]; ok {
				return app.Run(path)
			}
		}
	}

	return xdg.RunCmd("xdg-open", path)
}

var argPlaceholders = regexp.MustCompile("%[uUfF]")

func run(exec string, arg string, inTerminal bool) error {
	var argv = strings.Fields(exec)
	for i := 0; i < len(argv); i++ {
		argv[i] = argPlaceholders.ReplaceAllString(argv[i], arg)
	}

	// Get rid of empty arguments
	var left = 0
	for i := 0; i < len(argv); i++ {
		if len(strings.TrimSpace(argv[i])) > 0 {
			argv[left] = strings.TrimSpace(argv[i])
			left++
		}
	}
	argv = argv[0:left]

	if inTerminal {
		var terminal, ok = os.LookupEnv("TERMINAL")
		if !ok {
			return fmt.Errorf("Trying to run %s in terminal, but env variable TERMINAL not set", exec)
		}
		argv = append([]string{terminal, "-e"}, argv...)
	}

	return xdg.RunCmd(argv...)
}
