// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/searchutils"

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
	return &respond.StandardFormat{
		Self:         self,
		OnPost:       "Launch",
		OtherActions: self + "/actions",
		Type:         "application",
		Title:        d.Name,
		Comment:      d.Comment,
		IconName:     d.IconName,
		Data:         d,
		NoDisplay:    d.NoDisplay,
	}
}

func (d *DesktopApplication) collectActions(term string) respond.StandardFormatList {
	var sfl = make(respond.StandardFormatList, 0, len(d.DesktopActions))
	for _, act := range d.DesktopActions {
		if rank := searchutils.SimpleRank(act.Name, "", term); rank > -1 {
			sfl = append(sfl, act.ToStandardFormat().Ranked(rank))
		}
	}
	return sfl.SortByRank()
}

type DesktopAction struct {
	self     string
	Name     string
	Exec     string
	IconName string
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

func appSelf(appId string) string {
	if !strings.HasSuffix(appId, ".desktop") {
		log.Println("Weird application id:", appId)
		return ""
	} else {
		return "/application/" + appId[:len(appId)-8]
	}
}

type ApplicationMap map[string]*DesktopApplication
