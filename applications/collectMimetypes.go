package applications

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/surlykke/RefudeServices/icon"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/path"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/tr"
)

func collectMimetypes() map[string]*Mimetype {
	res := make(map[string]*Mimetype)

	for id, comment := range schemeHandlers {
		var mimetype, err = MakeMimetype(id)
		if err != nil {
			log.Warn("Problem making mimetype", id)
		} else {
			mimetype.Comment = comment
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
		log.Warn("Unable to open ", freedesktopOrgXml, ": ", err)
	}
	parseErr := xml.Unmarshal(xmlInput, &xmlCollector)
	if parseErr != nil {
		log.Warn("Error parsing: ", parseErr)
	}

	for _, tmp := range xmlCollector.MimeTypes {
		if mimeType, err := MakeMimetype(tmp.Type); err != nil {
			log.Warn(err)
		} else {
			for _, tmpComment := range tmp.Comment {
				if tr.LocaleMatch(tmpComment.Lang) || (tmpComment.Lang == "" && mimeType.Comment == "") {
					mimeType.Comment = tmpComment.Text
				}
			}

			for _, tmpAcronym := range tmp.Acronym {
				if tr.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && mimeType.Acronym == "") {
					mimeType.Acronym = tmpAcronym.Text
				}
			}

			for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
				if tr.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.ExpandedAcronym == "" {
					mimeType.ExpandedAcronym = tmpExpandedAcronym.Text
				}
			}

			if tmp.Icon.Name == "" {
				tmp.Icon.Name = strings.Replace(tmp.Type, "/", "-", -1)
			}
			mimeType.Icon = icon.Name(tmp.Icon.Name)

			for _, aliasStruct := range tmp.Alias {
				mimeType.Aliases = slice.AppendIfNotThere(mimeType.Aliases, aliasStruct.Type)
			}

			for _, tmpGlob := range tmp.Glob {
				mimeType.Globs = slice.AppendIfNotThere(mimeType.Globs, tmpGlob.Pattern)
			}

			for _, tmpSubClassOf := range tmp.SubClassOf {
				mimeType.SubClassOf = slice.AppendIfNotThere(mimeType.SubClassOf, tmpSubClassOf.Type)
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
		for i := 0; i < len(mt.SubClassOf); i++ {
			if ancestor, ok := res[mt.SubClassOf[i]]; ok {
				for _, id := range ancestor.SubClassOf {
					mt.SubClassOf = slice.AppendIfNotThere(mt.SubClassOf, id)
				}
			}
		}
	}

	for _, mt := range res {
		for _, aliasId := range mt.Aliases {
			if _, ok := res[aliasId]; !ok {
				var copy = *mt
				copy.Path = path.Of("/mimetype/", aliasId)
				copy.Id = aliasId
				copy.Aliases = []string{}
				res[aliasId] = &copy
			}
		}
	}

	return res

}
