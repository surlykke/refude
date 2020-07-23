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
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	self            string
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
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func MakeMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		return &Mimetype{
			self:        "/mimetype/" + id,
			Id:          id,
			Aliases:     []string{},
			Globs:       []string{},
			SubClassOf:  []string{},
			IconName:    "unknown",
			GenericIcon: "unknown",
		}, nil
	}
}

func (mt *Mimetype) ToStandardFormat() *respond.StandardFormat {
	return &respond.StandardFormat{
		Self:     mt.self,
		Type:     "mimetype",
		Title:    mt.Comment,
		Comment:  mt.Acronym,
		IconName: mt.IconName,
		Data:     mt,
	}
}

func (mt *Mimetype) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson2(w, mt.ToStandardFormat())
		/*} else if r.Method == "PATCH" {
		var decoder = json.NewDecoder(r.Body)
		var decoded = make(map[string]string)
		if err := decoder.Decode(&decoded); err != nil {
			respond.UnprocessableEntity(w, err)
		} else if defaultApp, ok := decoded["DefaultApp"]; !ok || len(decoded) != 1 {
			respond.UnprocessableEntity(w, fmt.Errorf("Patch payload should contain exactly one parameter: 'DefaultApp"))
		} else if err = setDefaultApp(mt.Id, defaultApp); err != nil {
			respond.ServerError(w, err)
		} else {
			respond.Accepted(w)
		}
		*/
	} else {
		respond.NotAllowed(w)
	}
}

func SetDefaultApp(mimetypeId string, appId string) error {
	if mt, ok := collectionStore.Load().(collection).mimetypes[mimetypeId]; ok {
		if mt.DefaultApp == appId {
			return nil
		}
	}
	path := xdg.ConfigHome + "/mimeapps.list"
	fmt.Println("reading", path)
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
		fmt.Println("Write", path)
		if err = xdg.WriteIniFile(path, iniFile); err != nil {
			return err
		}
		return nil
	}
}

func GetDefaultAppName(mimetypeId string) string {
	var c = collectionStore.Load().(collection)
	if mt, ok := c.mimetypes[mimetypeId]; ok {
		if mt.DefaultApp != "" {
			if app, ok := c.applications[mt.DefaultApp]; ok {
				return app.Name
			}
		}
	}

	return ""
}
