package mimetypes

import (
	"encoding/xml"
	"os"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/surlykke/RefudeServices/applications"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var handlerSubscription = applications.SubscribeToMimetypeHandlers()


func Run() {
	var mtRequests = repo.MakeAndRegisterRequestChan()
	var mimetypeRepo = repo.MakeRepo[*Mimetype]()
	var mtMap = collect()
	for _, mt := range mtMap {
		mimetypeRepo.Put(mt)
	}

	for {
		select {
		case req := <-mtRequests:
			mimetypeRepo.DoRequest(req)
		case handlers := <- handlerSubscription: 
			for mtId, apps := range handlers {
				var path = "/mimetype/" + mtId
				if mt, ok := mimetypeRepo.Get(path); ok {
					mt.Applications = apps
				}
			}
		}
	}

}

func watchDir(watcher *fsnotify.Watcher, dir string) {
	if xdg.DirOrFileExists(dir) {
		if err := watcher.Add(dir); err != nil {
			log.Warn("Could not watch:", dir, ":", err)
		}
	}
}

func collect() map[string]*Mimetype {
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
				if xdg.LocaleMatch(tmpComment.Lang) || (tmpComment.Lang == "" && mimeType.Comment == "") {
					mimeType.Comment = tmpComment.Text
				}
			}

			for _, tmpAcronym := range tmp.Acronym {
				if xdg.LocaleMatch(tmpAcronym.Lang) || (tmpAcronym.Lang == "" && mimeType.Acronym == "") {
					mimeType.Acronym = tmpAcronym.Text
				}
			}

			for _, tmpExpandedAcronym := range tmp.ExpandedAcronym {
				if xdg.LocaleMatch(tmpExpandedAcronym.Lang) || tmpExpandedAcronym.Lang == "" && mimeType.ExpandedAcronym == "" {
					mimeType.ExpandedAcronym = tmpExpandedAcronym.Text
				}
			}

			if tmp.Icon.Name != "" {
				mimeType.IconUrl = link.IconUrlFromName(tmp.Icon.Name)
			} else {
				mimeType.IconUrl = link.IconUrlFromName(strings.Replace(tmp.Type, "/", "-", -1))
			}

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
		for _, alias := range aliasTypes(mt) {
			if _, ok := res[alias.Path]; !ok {
				res[alias.Path] = alias
			}
		}
	}

	return res

	/* FIXME 	for mimetypeId, defaultAppIds := range defaultApps {
		if mimetype, ok := mimetypes[mimetypeId]; ok {
			mimetype.Applications = append(mimetype.Applications, defaultAppIds...)
		}
	}
	for appId, app := range apps {
		for _, mimetypeId := range app.Mimetypes {
			if mimetype, ok := mimetypes[mimetypeId]; ok {
				mimetype.Applications = slice.AppendIfNotThere(mimetype.Applications, appId)
			}
		}
	}*/

}

func aliasTypes(mt *Mimetype) []*Mimetype {
	var result = make([]*Mimetype, 0, len(mt.Aliases))
	for _, id := range mt.Aliases {
		var copy = *mt
		copy.Path = "/mimetype/" + id
		copy.Aliases = []string{}
		result = append(result, &copy)
	}

	return result
}


