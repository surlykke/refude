// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib"
	"golang.org/x/text/language"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

const MimetypeMediaType lib.MediaType = "application/vnd.org.refude.mimetype+json"

type Mimetype struct {
	lib.AbstractResource
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
		mt.Self = lib.Standardizef("/mimetypes/%s", id)
		mt.Mt = MimetypeMediaType


		if strings.HasPrefix(id, "x-scheme-handler/") {
			mt.Comment = id[len("x-scheme-handler/"):] + " url"
		} else {
			mt.Comment = id
		}

		return mt, nil
	}
}

func CollectMimeTypes() map[string]*Mimetype {
	xmlCollector := struct {
		XMLName xml.Name `xml:"mime-info"`
		MimeTypes []struct {
			Type string `xml:"type,attr"`
			Comment []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"comment"`
			Acronym []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"acronym"`
			ExpandedAcronym []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"expanded-acronym"`
			Alias []struct {
				Type string `xml:"type,attr"`
			} `xml:"alias"`
			Glob []struct {
				Pattern string `xml:"pattern,attr"`
			} `xml:"glob"`
			SubClassOf []struct {
				Type string `xml:"type,attr"`
			} `xml:"sub-class-of"`
			Icon struct {
				Name string `xml:"name,attr"`
			} `xml:"icon"`
			GenericIcon struct {
				Name string `xml:"name,attr"`
			} `xml:"generic-icon"`
		} `xml:"mime-type"`
	}{}

	xmlInput, err := ioutil.ReadFile(freedesktopOrgXml)
	if err != nil {
		fmt.Println("Unable to open ", freedesktopOrgXml, ": ", err)
	}
	parseErr := xml.Unmarshal(xmlInput, &xmlCollector)
	if parseErr != nil {
		fmt.Println("Error parsing: ", parseErr)
	}

	res := make(map[string]*Mimetype)
	for _, tmp := range xmlCollector.MimeTypes {
		mimeType, err := NewMimetype(tmp.Type)
		var collectedLocales = make(map[string]bool)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, tmpComment := range tmp.Comment {
			if lib.LocaleMatch(tmpComment.Lang) || (tmpComment.Lang == "" && mimeType.Comment == "") {
				mimeType.Comment = tmpComment.Text
			}
		}

		for _, tmpAcronym := range tmp.Acronym {
			if lib.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && mimeType.Acronym == "") {
				mimeType.Acronym = tmpAcronym.Text
			}
		}

		for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
			if (lib.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.Acronym == "") {
				mimeType.ExpandedAcronym = tmpExpandedAcronym.Text
			}
		}

		if tmp.Icon.Name != "" {
			mimeType.IconName = tmp.Icon.Name
		} else {
			mimeType.IconName = strings.Replace(mimeType.Id, "/", "-", -1)
		}

		for _, aliasStruct := range tmp.Alias {
			mimeType.Aliases = lib.AppendIfNotThere(mimeType.Aliases, aliasStruct.Type)
		}

		for _, tmpGlob := range tmp.Glob {
			mimeType.Globs = lib.AppendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
		}

		for _, tmpSubClassOf := range tmp.SubClassOf {
			mimeType.SubClassOf = lib.AppendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
		}

		if tmp.GenericIcon.Name != "" {
			mimeType.GenericIcon = tmp.GenericIcon.Name
		} else {
			slashPos := strings.Index(mimeType.Id, "/")
			mimeType.GenericIcon = mimeType.Id[:slashPos] + "-x-generic"
		}

		var tags = make([]language.Tag, len(collectedLocales))
		for locale, _ := range collectedLocales {
			tags = append(tags, language.Make(locale))
		}

	    res[mimeType.Id] = mimeType
	}

	return res
}
