// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/icons"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	respond.Resource
	Id              string
	Comment         string
	Acronym         string `json:",omitempty"`
	ExpandedAcronym string `json:",omitempty"`
	Aliases         []string
	Globs           []string
	SubClassOf      []string
	IconName        string
	GenericIcon     string
	DefaultApp      string `json:",omitempty"`
	DefaultAppPath  string `json:",omitempty"`
	path            string
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func MakeMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		var mt = Mimetype{
			Id:          id,
			Aliases:     []string{},
			Globs:       []string{},
			SubClassOf:  []string{},
			IconName:    "unknown",
			GenericIcon: "unknown",
			path:        "/mimetype/" + id,
		}
		return &mt, nil
	}
}

func SetDefaultApp(mimetypeId string, appId string) error {
	if mt, ok := collectionStore.Load().(collection).mimetypes[mimetypeId]; ok {
		if mt.DefaultApp == appId {
			return nil
		}
	}
	path := xdg.ConfigHome + "/mimeapps.list"
	if iniFile, err := xdg.ReadIniFile(path); err != nil && !os.IsNotExist(err) {
		return err
	} else {
		var defaultGroup = iniFile.FindGroup("Default Applications")
		if defaultGroup == nil {
			defaultGroup = &xdg.Group{Name: "Default Applications", Entries: make(map[string]string)}
			iniFile = append(iniFile, defaultGroup)
		}
		var defaultAppsS = defaultGroup.Entries[mimetypeId]
		var defaultApps = slice.Split(defaultAppsS, ";")
		defaultApps = slice.PushFront(appId, slice.Remove(defaultApps, appId))
		defaultAppsS = strings.Join(defaultApps, ";")
		defaultGroup.Entries[mimetypeId] = defaultAppsS
		if err = xdg.WriteIniFile(path, iniFile); err != nil {
			return err
		}
		return nil
	}
}

func GetDefaultApp(mimetypeId string) string {
	var c = collectionStore.Load().(collection)
	if mt, ok := c.mimetypes[mimetypeId]; ok {
		if mt.DefaultApp != "" {
			if app, ok := c.applications[mt.DefaultApp]; ok {
				return app.Id
			}
		}
	}

	return ""
}

func IconForMimetype(mimetypeId string) string {
	return icons.IconUrl(strings.ReplaceAll(mimetypeId, "/", "-"))
}
