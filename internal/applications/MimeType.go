// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package applications

import (
	"os"
	"regexp"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/xdg"
)

var freedesktopOrgXml = ""

func init() {
	for _, dir := range xdg.DataDirs {
		if _, err := os.Stat(dir + "/mime/packages/freedesktop.org.xml"); err == nil {
			freedesktopOrgXml = dir + "/mime/packages/freedesktop.org.xml"
			return
		}
	}
}

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
			Base:        *entity.MakeBase("", "", "", "Mimetype"),
			Id:          id,
			Aliases:     []string{},
			Globs:       []string{},
			SubClassOf:  []string{},
			GenericIcon: "unknown",
		}
		return &mt, nil
	}
}
