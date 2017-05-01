/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (

	"encoding/xml"
	"github.com/surlykke/RefudeServices/common"
	"io/ioutil"
	"fmt"
	"regexp"
	"strings"
	"net/http"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	Type                   string
	Subtype                string
	Comment                string
	Acronym                string
	ExpandedAcronym        string
	Aliases                common.StringList
	Globs                  common.StringList
	SubClassOf             common.StringList
	IconName               string
	GenericIcon            string
	AssociatedApplications common.StringList
	DefaultApplications    []string
}

func (mt *Mimetype) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		common.ServeAsJson(w, r, mt)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}


func CollectMimeTypes() map[string]*Mimetype {
	xmlCollector := struct {
		XMLName   xml.Name `xml:"mime-info"`
		MimeTypes []struct {
			Type    string `xml:"type,attr"`
			Comment []struct {
				Lang string `xml:"lang,attr"`
				Text string `xml:",chardata"`
			} `xml:"comment"`
			Acronym         string `xml:"acronym"`
			ExpandedAcronym string `xml:"expanded-acronym"`
			Alias           []struct {
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

	typePattern, err := regexp.Compile(`^([^/]+)/([^/]+)$`)
	if err != nil {
		panic(err)
	}

	res := make(map[string]*Mimetype)
	for _, tmp := range xmlCollector.MimeTypes {
		mimeType := Mimetype{}

		typeElements := typePattern.FindStringSubmatch(tmp.Type)
		if len(typeElements) == 3 {
			mimeType.Type = typeElements[1]
			mimeType.Subtype = typeElements[2]
		} else {
			fmt.Println("Incomprehensible mimetype: ", tmp.Type)
			continue
		}

		for _, tmpComment := range tmp.Comment {
			if tmpComment.Lang == "" {
				mimeType.Comment = tmpComment.Text // FIXME
			}
		}

		mimeType.Acronym = tmp.Acronym
		mimeType.ExpandedAcronym = tmp.ExpandedAcronym

		mimeType.Aliases = make(common.StringList, 0)
		for _, aliasStruct := range tmp.Alias {
			mimeType.Aliases = common.AppendIfNotThere(mimeType.Aliases, aliasStruct.Type)
		}

		mimeType.Globs = make(common.StringList, 0)
		for _, tmpGlob := range tmp.Glob {
			mimeType.Globs = common.AppendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
		}

		mimeType.SubClassOf = make(common.StringList, 0)
		for _, tmpSubClassOf := range tmp.SubClassOf {
			mimeType.SubClassOf = common.AppendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
		}

		if tmp.Icon.Name != "" {
			mimeType.IconName = tmp.Icon.Name
		} else {
			mimeType.IconName = strings.Replace(mimeType.Type, "/", "-", -1)
		}

		if tmp.GenericIcon.Name != "" {
			mimeType.GenericIcon = tmp.GenericIcon.Name
		} else {
			mimeType.GenericIcon = mimeType.Type + "-x-generic"
		}

		mimeType.AssociatedApplications = make(common.StringList, 0)
		mimeType.DefaultApplications = make([]string, 0)

		res[mimeType.Type+"/"+mimeType.Subtype] = &mimeType
	}

	return res
}
