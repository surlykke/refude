// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"net/http"
	"golang.org/x/text/language"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/query"
)

const DesktopApplicationMediaType mediatype.MediaType = "application/vnd.org.refude.desktopapplication+json"

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
}

type Action struct {
	Name     string
	Exec     string
	IconName string
	IconPath string
	IconUrl  string
}

type IconPath struct {
	path string
}

func (ip IconPath) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		http.ServeFile(w, r, ip.path)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (ip IconPath) Match(m query.Matcher) bool {
	return false
}

func (ip IconPath) Mt() mediatype.MediaType {
	return mediatype.MediaType("image/png")
}

