// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (

	"encoding/xml"
	"io/ioutil"
	"fmt"
	"regexp"
	"strings"
	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib/utils"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/surlykke/RefudeServices/lib/ini"
	"github.com/surlykke/RefudeServices/lib/resource"
	"golang.org/x/text/language"
)

const freedesktopOrgXml = "/usr/share/mime/packages/freedesktop.org.xml"

type Mimetype struct {
	Id                     string
	Comment                localizedString
	Acronym                localizedString
	ExpandedAcronym        localizedString
	Aliases                []string
	Globs                  []string
	SubClassOf             []string
	IconName               string
	GenericIcon            string
	AssociatedApplications []string
	DefaultApplications    []string
	languages              language.Matcher
}

type LocalizedMimetype struct {
	Id                     string
	Comment                string
	Acronym                string
	ExpandedAcronym        string
	Aliases                []string
	Globs                  []string
	SubClassOf             []string
	IconName               string
	GenericIcon            string
	AssociatedApplications []string
	DefaultApplications    []string
}

func (mt *Mimetype) localize(locale string) *LocalizedMimetype {
	return &LocalizedMimetype{
		Id: mt.Id,
		Comment: mt.Comment.localize(locale),
		Acronym: mt.Acronym.localize(locale),
		ExpandedAcronym: mt.ExpandedAcronym.localize(locale),
		Aliases: mt.Aliases,
		Globs: mt.Globs,
		SubClassOf: mt.SubClassOf,
		IconName: mt.IconName,
		GenericIcon: mt.GenericIcon,
		AssociatedApplications: mt.AssociatedApplications,
		DefaultApplications: mt.DefaultApplications,
	}
}


func (mt *Mimetype) GET(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Incoming: ", r)
	var locale = getPreferredLocale(r, mt.languages)
	resource.JsonGET(mt.localize(locale), w)
}

func (mt *Mimetype) POST(w http.ResponseWriter, r *http.Request) {
	defaultAppId := r.URL.Query()["defaultApp"]
	if len(defaultAppId) != 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	mimetypeId := mt.Id
	appId := defaultAppId[0]

	fmt.Println("Setting default application: ", mimetypeId, " -> ", appId)

	path := xdg.ConfigHome + "/mimeapps.list"

	if iniFile, err := ini.ReadIniFile(path); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		defaultApplications := utils.Split(iniFile.Value("Default Applications", mimetypeId), ";")
		defaultApplications = utils.Remove(defaultApplications, appId)
		defaultApplications = utils.PushFront(appId, defaultApplications)

		fmt.Println("Setting: ", mimetypeId, strings.Join(defaultApplications, ";"))
		iniFile.SetValue("Default Applications", mimetypeId, strings.Join(defaultApplications, ";"))
		if err = ini.WriteIniFile(path, iniFile); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

var	mimetypePattern = func() *regexp.Regexp {
	pattern, err := regexp.Compile(`^([^/]+)/([^/]+)$`)
	if err != nil {
		panic(err)
	}
	return pattern
}()

func NewMimetype(id string) (*Mimetype, error) {

	if ! mimetypePattern.MatchString(id) {
		return nil, errors.New("Incomprehensible mimetype: " + id)
	} else {
		mt := &Mimetype{
			Id:                     id,
			Comment:                make(localizedString),
			Acronym:                make(localizedString),
			ExpandedAcronym:        make(localizedString),
			Aliases:                make([]string, 0),
			Globs:                  make([]string, 0),
			SubClassOf:             make([]string, 0),
			IconName:               "unknown",
			GenericIcon:            "unknown",
			AssociatedApplications: make([]string, 0),
			DefaultApplications:    make([]string, 0),
		}
		if strings.HasPrefix(id, "x-scheme-handler/") {
			mt.Comment[""] = id[len("x-scheme-handler/"):] + " url"
		} else {
			mt.Comment[""] = id
		}

		return mt, nil
	}
}



func CollectMimeTypes() map[string]*Mimetype {
	var allCollectedLocales = make(map[string]bool)
	xmlCollector := struct {
		XMLName   xml.Name `xml:"mime-info"`
		MimeTypes []struct {
			Type    string `xml:"type,attr"`
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

	res := make(map[string]*Mimetype)
	for _, tmp := range xmlCollector.MimeTypes {
		mimeType, err := NewMimetype(tmp.Type)
		var collectedLocales = make(map[string]bool)
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, tmpComment := range tmp.Comment {
			locale := transformLanguageTag(tmpComment.Lang)
			mimeType.Comment[locale] = tmpComment.Text
			collectedLocales[locale] = true
			allCollectedLocales[locale] = true
		}

		for _, tmpAcronym := range tmp.Acronym {
			locale := transformLanguageTag(tmpAcronym.Lang)
			mimeType.Acronym[locale] = tmpAcronym.Text
			collectedLocales[locale] = true
			allCollectedLocales[locale] = true
		}

		for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
			locale := transformLanguageTag(tmpExpandedAcronym.Lang)
			mimeType.ExpandedAcronym[locale] = tmpExpandedAcronym.Text
			collectedLocales[locale] = true
			allCollectedLocales[locale] = true
		}

		if tmp.Icon.Name != "" {
			mimeType.IconName = tmp.Icon.Name
		} else {
			mimeType.IconName = strings.Replace(mimeType.Id, "/", "-", -1)
		}

		for _, aliasStruct := range tmp.Alias {
			mimeType.Aliases = utils.AppendIfNotThere(mimeType.Aliases, aliasStruct.Type)
		}

		for _, tmpGlob := range tmp.Glob {
			mimeType.Globs = utils.AppendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
		}

		for _, tmpSubClassOf := range tmp.SubClassOf {
			mimeType.SubClassOf = utils.AppendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
		}


		if tmp.GenericIcon.Name != "" {
			mimeType.GenericIcon = tmp.GenericIcon.Name
		} else {
			slashPos := strings.Index(mimeType.Id, "/")
			mimeType.GenericIcon = mimeType.Id[:slashPos] + "-x-generic"
		}

		var tags = make([]language.Tag, len(collectedLocales))
		for locale,_ := range collectedLocales {
			tags = append(tags, language.Make(locale))
		}
		mimeType.languages = language.NewMatcher(tags)

		res[mimeType.Id] = mimeType
	}

	fmt.Println("Collected locales: ")
	for locale, _ := range allCollectedLocales {
		fmt.Print(locale, " ")
	}
	fmt.Println()
	return res
}

