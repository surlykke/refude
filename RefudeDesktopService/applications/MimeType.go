// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"

	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	resource.Links
	resource.Actions
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
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func NewMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		mt := &Mimetype{}
		mt.Self = mimetypeSelf(id)
		mt.RefudeType = "mimetype"
		mt.Id = id
		mt.Aliases = []string{}
		mt.Globs = []string{}
		mt.SubClassOf = []string{}
		mt.IconName = "unknown"
		mt.GenericIcon = "unknown"

		if strings.HasPrefix(id, "x-scheme-handler/") {
			mt.Comment = id[len("x-scheme-handler/"):] + " url"
		} else {
			mt.Comment = id
		}

		return mt, nil
	}
}

func (mt *Mimetype) PATCH(w http.ResponseWriter, r *http.Request) {
	var decoder = json.NewDecoder(r.Body)
	var decoded = make(map[string]string)
	if err := decoder.Decode(&decoded); err != nil {
		requests.ReportUnprocessableEntity(w, err)
	} else if defaultApp, ok := decoded["DefaultApp"]; !ok || len(decoded) != 1 {
		requests.ReportUnprocessableEntity(w, fmt.Errorf("Patch payload should contain exactly one parameter: 'DefaultApp"))
	} else if err = setDefaultApp(mt.Id, defaultApp); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func setDefaultApp(mimetypeId string, appId string) error {
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

func mimetypeSelf(mimetypeId string) string {
	return fmt.Sprintf("/mimetype/%s", mimetypeId)
}
