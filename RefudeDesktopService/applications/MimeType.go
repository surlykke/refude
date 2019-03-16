// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package applications

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

const MimetypeMediaType resource.MediaType = "application/vnd.org.refude.mimetype+json"

type Mimetype struct {
	resource.AbstractResource
	Id              string
	Comment         string
	Acronym         string `json:",omitempty"`
	ExpandedAcronym string `json:",omitempty"`
	Aliases         []string
	Globs           []string
	SubClassOf      []string
	IconName        string
	GenericIcon     string
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func NewMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		mt := &Mimetype{
			Id:          id,
			Aliases:     []string{},
			Globs:       []string{},
			SubClassOf:  []string{},
			IconName:    "unknown",
			GenericIcon: "unknown",
		}
		mt.AbstractResource = resource.MakeAbstractResource(mimetypeSelf(id), MimetypeMediaType)

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
		log.Println(fmt.Sprint(err))
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
			log.Println(fmt.Sprint(err))
		}
	}
}

type MimetypeCollection struct {
	mutex sync.Mutex
	mimetypes map[resource.StandardizedPath]*Mimetype
	server.CachingJsonGetter
}

func MakeMimetypecollection() *MimetypeCollection {
	var mc = &MimetypeCollection{}
	mc.mimetypes = make(map[resource.StandardizedPath]*Mimetype)
	mc.CachingJsonGetter = server.MakeCachingJsonGetter(mc)
	return mc
}

func (mc *MimetypeCollection) PATCH(w http.ResponseWriter, r *http.Request) {
	// FIXME
}

func (mc *MimetypeCollection) GetSingle(r *http.Request) interface{} {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	if mt, ok := mc.mimetypes[resource.Standardize(r.URL.Path)]; ok {
		return mt
	} else {
		return nil
	}
}

func (dac *MimetypeCollection) GetCollection(r *http.Request) []interface{} {
	dac.mutex.Lock()
	defer dac.mutex.Unlock()

	if r.URL.Path == "/mimetypes" {
		var result = make([]interface{}, 0, len(dac.mimetypes))
		for _, app := range dac.mimetypes {
			result = append(result, app)
		}
		return result
	} else {
		return nil
	}
}

func mimetypeSelf(mimetypeId string) resource.StandardizedPath {
	return resource.Standardizef("/mimetype/%s", mimetypeId)
}
