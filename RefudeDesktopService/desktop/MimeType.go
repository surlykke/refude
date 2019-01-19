// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package desktop

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

const MimetypeMediaType resource.MediaType = "application/vnd.org.refude.mimetype+json"

type Mimetype struct {
	resource.AbstractResource
	Id                     string
	Comment                string
	Acronym                string `json:",omitempty"`
	ExpandedAcronym        string `json:",omitempty"`
	Aliases                []string
	Globs                  []string
	SubClassOf             []string
	IconName               string
	GenericIcon            string
	AssociatedApplications []string
	DefaultApplication     string `json:",omitempty"`
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func NewMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		mt := &Mimetype{
			Id:           id,
			Aliases:      []string{},
			Globs:        []string{},
			SubClassOf:   []string{},
			IconName:     "unknown",
			GenericIcon:  "unknown",
		}
		mt.AbstractResource = resource.MakeAbstractResource(resource.Standardizef("/mimetypes/%s", id), MimetypeMediaType)

		if strings.HasPrefix(id, "x-scheme-handler/") {
			mt.Comment = id[len("x-scheme-handler/"):] + " url"
		} else {
			mt.Comment = id
		}

		return mt, nil
	}
}

func (mt *Mimetype) POST(w http.ResponseWriter, r *http.Request) {
	defaultAppId := r.URL.Query()["defaultApp"]
	if len(defaultAppId) != 1 || defaultAppId[0] == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		go setDefaultApp(mt.Id, defaultAppId[0])
		w.WriteHeader(http.StatusAccepted)
	}
}

func setDefaultApp(mimetypeId string, appId string) {
	path := xdg.ConfigHome + "/mimeapps.list"

	if iniFile, err := xdg.ReadIniFile(path); err != nil && !os.IsNotExist(err) {
		reportError(fmt.Sprint(err))
	} else {
		var defaultGroup = iniFile.FindGroup("Default Applications")
		if defaultGroup == nil {
			defaultGroup = &xdg.Group{"Default Applications", make(map[string]string)}
			iniFile = append(iniFile, defaultGroup)
		}
		var defaultAppsS = defaultGroup.Entries[mimetypeId]
		var defaultApps = slice.Split(defaultAppsS, ";")
		defaultApps = slice.PushFront(appId, slice.Remove(defaultApps, appId))
		defaultAppsS = strings.Join(defaultApps, ";")
		defaultGroup.Entries[mimetypeId] = defaultAppsS
		if err = xdg.WriteIniFile(path, iniFile); err != nil {
			reportError(fmt.Sprint(err))
		}
	}
}

