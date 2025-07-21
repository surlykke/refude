// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"regexp"

	"github.com/surlykke/RefudeServices/server/lib/entity"
	"github.com/surlykke/RefudeServices/server/lib/mediatype"

	"github.com/pkg/errors"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	entity.Base
	Id              string
	Comment         string
	Acronym         string `json:",omitempty"`
	ExpandedAcronym string `json:",omitempty"`
	Aliases         []string
	Globs           []string
	SubClassOf      []string
	GenericIcon     string
	Applications    []string
}

var mimetypePattern = regexp.MustCompile(`^([^/]+)/([^/]+)$`)

func MakeMimetype(id string) (*Mimetype, error) {

	if !mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		var mt = Mimetype{
			Base:        *entity.MakeBase("", "", "", mediatype.Mimetype),
			Id:          id,
			Aliases:     []string{},
			Globs:       []string{},
			SubClassOf:  []string{},
			GenericIcon: "unknown",
		}
		return &mt, nil
	}
}
