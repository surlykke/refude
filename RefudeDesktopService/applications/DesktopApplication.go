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

	"github.com/surlykke/RefudeServices/lib/respond"

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

func (d *DesktopApplication) ToStandardFormat() *respond.StandardFormat {
	var self = appSelf(d.Id)
	var otherActions string
	if len(d.DesktopActions) > 0 {
		otherActions = self + "/actions"
	}
	return &respond.StandardFormat{
		Self:         self,
		OnPost:       "Launch",
		OtherActions: otherActions,
		Type:         "application",
		Title:        d.Name,
		Comment:      d.Comment,
		IconName:     d.IconName,
		Data:         d,
		NoDisplay:    d.NoDisplay,
	}
}

func (d *DesktopApplication) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson2(w, d.ToStandardFormat())
	} else if r.Method == "POST" {
		respond.AcceptedAndThen(w, func() { d.Run("") })
	} else {
		respond.NotAllowed(w)
	}
}

func (d *DesktopApplication) collectActions(term string) respond.StandardFormatList {
	var sfl = make(respond.StandardFormatList, 0, len(d.DesktopActions))
	for _, act := range d.DesktopActions {
		sfl = append(sfl, act.ToStandardFormat())
	}
	return sfl
}

func (d *DesktopApplication) Run(arg string) error {
	return run(d.Exec, arg, d.Terminal)
}

func appSelf(appId string) string {
	return "/application/" + appId
}

type DesktopAction struct {
	self     string
	Name     string
	Exec     string
	IconName string
}

func (da *DesktopAction) Run(arg string) error {
	return run(da.Exec, arg, false)
}

func (da *DesktopAction) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     da.self,
		OnPost:   "Launch",
		Type:     "applicationaction",
		Title:    da.Name,
		IconName: da.IconName,
		Data:     da,
	}
}

func (da *DesktopAction) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson2(w, da.ToStandardFormat())
	} else if r.Method == "POST" {
		respond.AcceptedAndThen(w, func() { da.Run("") })
	} else {
		respond.NotAllowed(w)
	}
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
