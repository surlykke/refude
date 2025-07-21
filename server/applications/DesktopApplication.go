// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/server/lib/entity"
	"github.com/surlykke/RefudeServices/server/lib/icon"
	"github.com/surlykke/RefudeServices/server/lib/response"

	"github.com/surlykke/RefudeServices/server/lib/xdg"
)

type DesktopApplication struct {
	entity.Base
	Comment         string `json:",omitempty"`
	Type            string
	Version         string `json:",omitempty"`
	GenericName     string `json:",omitempty"`
	NoDisplay       bool
	Hidden          bool
	OnlyShowIn      []string
	NotShowIn       []string
	DbusActivatable bool   `json:",omitempty"`
	TryExec         string `json:",omitempty"`
	Exec            string `json:",omitempty"`
	WorkingDir      string //`json:",omitempty"`
	Terminal        bool
	Categories      []string
	Implements      []string
	StartupNotify   bool
	StartupWmClass  string `json:",omitempty"`
	Url             string `json:",omitempty"`
	DesktopActions  []DesktopAction
	DesktopId       string
	Mimetypes       []string
	DesktopFile     string
}

func (d *DesktopApplication) OmitFromSearch() bool {
	return d.NoDisplay
}

func (d *DesktopApplication) Run(arg string) error {
	return run(d.Exec, arg, d.Terminal)
}

type DesktopAction struct {
	id   string
	Name string
	Exec string
	Icon icon.Name
}

func (d *DesktopApplication) DoPost(action string) response.Response {
	if action == "" {
		return postHelper(d.Exec, d.Terminal)
	} else {
		for _, dac := range d.DesktopActions {
			if action == dac.id {
				return postHelper(dac.Exec, d.Terminal)
			}
		}
	}
	return response.NotFound()
}

func postHelper(exec string, terminal bool) response.Response {
	if err := run(exec, "", terminal); err != nil {
		return response.ServerError(err)
	} else {
		return response.Accepted()
	}
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
