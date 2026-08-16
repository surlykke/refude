// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package applications

import (
	"encoding/xml"
	"log"
	"os"
	"strings"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/translate"
)

func collectMimetypes() map[string]Mimetype {
	res := make(map[string]Mimetype)

	for id, _ := range schemeHandlers {
		var mimetype, err = MakeMimetype(id)
		if err != nil {
			log.Print("Problem making mimetype", id)
		} else {
			mimetype.Title = strings.Split(id, "/")[1] + " handler"
			mimetype.Path = "/mimetype/" + id
			res[id] = mimetype
		}
	}

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

	xmlInput, err := os.ReadFile(freedesktopOrgXml)
	if err != nil {
		log.Print("Unable to open ", freedesktopOrgXml, ": ", err)
	}
	parseErr := xml.Unmarshal(xmlInput, &xmlCollector)
	if parseErr != nil {
		log.Print("Error parsing: ", parseErr)
	}

	for _, tmp := range xmlCollector.MimeTypes {
		if !mimetypePattern.MatchString(tmp.Type) {
			log.Print("Incomprehensible mimetype:", tmp.Type)
		} else {
			var acronym = ""
			var comment = ""
			var expandedAcronym string
			var iconName string = ""

			for _, tmpAcronym := range tmp.Acronym {
				if translate.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && acronym == "") {
					acronym = tmpAcronym.Text
				}
			}

			for _, tmpComment := range tmp.Comment {
				if translate.LocaleMatch(tmpComment.Lang) || (tmpComment.Lang == "" && comment == "") {
					comment = tmpComment.Text
				}
			}

			for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
				if translate.LocaleMatch(tmpExpandedAcronym.Lang) || (tmpExpandedAcronym.Lang == "" && expandedAcronym == "") {
					expandedAcronym = tmpExpandedAcronym.Text
				}
			}

			if tmp.Icon.Name == "" {
				tmp.Icon.Name = strings.ReplaceAll(tmp.Type, "/", "-")
			}
			iconName = tmp.Icon.Name

			var mimeType = Mimetype{Base: resource.MakeBase("/mimetype/" + tmp.Type, acronym, "Mimetype", iconName, "Mimetype"), Id: tmp.Type}
			mimeType.Path = "/mimetype/" + mimeType.Id 
			for _, tmpAcronym := range tmp.Acronym {
				if translate.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && mimeType.Acronym == "") {
					mimeType.Acronym = tmpAcronym.Text
				}
			}

			for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
				if translate.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.ExpandedAcronym == "" {
					mimeType.ExpandedAcronym = tmpExpandedAcronym.Text
				}
			}

			for _, aliasStruct := range tmp.Alias {
				mimeType.Aliases = appendIfNotThere(mimeType.Aliases, aliasStruct.Type)
			}

			for _, tmpGlob := range tmp.Glob {
				mimeType.Globs = appendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
			}

			for _, tmpSubClassOf := range tmp.SubClassOf {
				mimeType.SubClassOf = appendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
			}

			if tmp.GenericIcon.Name != "" {
				mimeType.GenericIcon = tmp.GenericIcon.Name
			} else {
				slashPos := strings.Index(tmp.Type, "/")
				mimeType.GenericIcon = tmp.Type[:slashPos] + "-x-generic"
			}

			res[mimeType.Id] = mimeType
		}
	}

	// Do a transitive closure on 'SubClassOf'
	for _, mt := range res {
		for i := range mt.SubClassOf {
			if ancestor, ok := res[mt.SubClassOf[i]]; ok {
				for _, id := range ancestor.SubClassOf {
					mt.SubClassOf = appendIfNotThere(mt.SubClassOf, id)
				}
			}
		}
	}

	for _, mt := range res {
		for _, aliasId := range mt.Aliases {
			if _, ok := res[aliasId]; !ok {
				var copy = mt
				copy.Base = resource.MakeBase(mimeTypeId2path(aliasId), mt.Title, mt.Summary, mt.Icon, mt.Keywords...)
				copy.Id = aliasId
				copy.Aliases = []string{}
				res[aliasId] = copy
			}
		}
	}
	return res

}
