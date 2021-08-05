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

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
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
	Icon            string `json:",omitempty"`
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
	path            string `json:"-"`
}

func (d *DesktopApplication) Run(arg string) error {
	return run(d.Exec, arg, d.Terminal)
}

func (d *DesktopApplication) Links() []resource.Link {
	var links = []resource.Link{
		resource.MakeLink(d.self, d.Name, d.Icon, relation.Self),
		resource.MakeLink(d.self, "Launch", "", relation.DefaultAction),
	}
	for _, da := range d.DesktopActions {
		links = append(links, resource.MakeLink(d.self+"?action="+da.id, da.Name, da.Icon, relation.Action))
	}
	return links

}

func (d *DesktopApplication) RefudeType() string {
	return "application"
}

type DesktopAction struct {
	id   string
	Name string
	Exec string
	Icon string
}

func (da *DesktopAction) Run(arg string) error {
	return run(da.Exec, arg, false)
}

func (d *DesktopApplication) DoPost(w http.ResponseWriter, r *http.Request) {
	var exec string
	var terminal bool
	var action = requests.GetSingleQueryParameter(r, "action", "")
	if action == "" {
		exec, terminal = d.Exec, d.Terminal
	} else {
		for _, da := range d.DesktopActions {
			if action == da.id {
				exec = da.Exec
			}
		}
	}
	if exec != "" {
		if err := run(exec, "", terminal); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
	} else {
		respond.NotFound(w)
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
			return fmt.Errorf("trying to run %s in terminal, but env variable TERMINAL not set", exec)
		}
		argv = append([]string{terminal, "-e"}, argv...)
	}

	return xdg.RunCmd(argv...)
}
