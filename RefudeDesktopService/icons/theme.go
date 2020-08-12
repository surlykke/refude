// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type IconTheme struct {
	respond.Links `json:"_links"`
	Id            string
	Name          string
	Comment       string
	Inherits      []string
	Dirs          map[string]IconDir
}

func (it *IconTheme) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		respond.AsJson(w, it)
	} else {
		respond.NotAllowed(w)
	}
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

func readTheme(indexThemeFilePath string) (*IconTheme, bool) {
	// id of a theme is the name of the directory in which its index.theme file resides
	var themeId = filepath.Base(filepath.Dir(indexThemeFilePath))

	if themeId == "." || themeId == ".." || themeId == "/" {
		log.Println("Could not figure theme id from path:", indexThemeFilePath)
	}

	iniFile, err := xdg.ReadIniFile(indexThemeFilePath)
	if err != nil {
		log.Println("Error reading theme:", err)
		return nil, false
	}

	if len(iniFile) < 1 || iniFile[0].Name != "Icon Theme" {
		log.Printf("Error reading %s , expected 'Icon Theme' at start", indexThemeFilePath)
		return nil, false
	}

	themeGroup := iniFile[0]

	theme := IconTheme{}
	theme.Id = themeId
	theme.Name = themeGroup.Entries["Name"]
	theme.Comment = themeGroup.Entries["Comment"]
	theme.Inherits = slice.Split(themeGroup.Entries["Inherits"], ",")
	theme.Dirs = make(map[string]IconDir)
	directories := slice.Split(themeGroup.Entries["Directories"], ",")
	if len(directories) == 0 {
		log.Printf("Ignoring theme %s - no directories", theme.Id)
		return nil, false
	}
	for _, iniGroup := range iniFile[1:] {

		if !slice.Contains(directories, iniGroup.Name) {
			//fmt.Fprintln(os.Stderr, iniGroup.Name, " not found in Directories")
			continue
		}

		size, sizeGiven := readUint32(iniGroup.Entries["Size"])
		if !sizeGiven {
			//fmt.Fprintln(os.Stderr, "Skipping ", iniGroup.Name, " - no size given")
			continue
		}

		minSize, minSizeGiven := readUint32(iniGroup.Entries["MinSize"])
		maxSize, maxSizeGiven := readUint32(iniGroup.Entries["MaxSize"])
		threshold := readUint32OrFallback(iniGroup.Entries["Threshold"], 2)
		sizeType := iniGroup.Entries["Type"]
		if strings.EqualFold(sizeType, "Fixed") {
			minSize = size
			maxSize = size
		} else if strings.EqualFold(sizeType, "Scalable") {
			if !minSizeGiven {
				minSize = size
			}
			if !maxSizeGiven {
				maxSize = size
			}
		} else if strings.EqualFold(sizeType, "Threshold") {
			minSize = size - threshold
			maxSize = size + threshold
		} else {
			_, _ = fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name, ", type must be given as 'Fixed', 'Scalable' or 'Threshold', was: ", sizeType)
			continue
		}

		theme.Dirs[iniGroup.Name] = IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entries["Context"]}
	}

	theme.Links = respond.Links{{Href: "/icontheme/" + theme.Id, Rel: respond.Self, Profile: "/profile/icontheme"}}
	return &theme, true
}

func readUint32(uintAsString string) (uint32, bool) {
	res, err := strconv.ParseUint(uintAsString, 10, 32)
	return uint32(res), err == nil
}

func readUint32OrFallback(uintAsString string, fallback uint32) uint32 {
	if res, ok := readUint32(uintAsString); ok {
		return res
	} else {
		return fallback
	}
}
