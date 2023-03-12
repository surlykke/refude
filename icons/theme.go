// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type IconTheme struct {
	ThemeId  string
	self     string
	Name     string
	Comment  string
	Inherits []string
	Dirs     []IconDir
}

func (it *IconTheme) Path() string {
	return it.ThemeId
}

func (it *IconTheme) Presentation() (title string, comment string, icon link.Href, profile string) {
	return it.Name, it.Comment, "", "icontheme"
}

func (it *IconTheme) Links(self, term string) link.List {
	return link.List{}
}

func (it *IconTheme) RelevantForSearch() bool {
	return true
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

func readThemes() map[string]*IconTheme {
	var themeMap = make(map[string]*IconTheme)

	for _, basedir := range basedirs {
		if indexFilePaths, err := filepath.Glob(basedir + "/*/index.theme"); err != nil {
			log.Info("Could not look for index.theme files:", err)
		} else {
			for _, indexFilePath := range indexFilePaths {
				if theme, ok := readTheme(indexFilePath); !ok {
					log.Warn("Could not read", indexFilePath)
				} else if _, ok := themeMap[theme.ThemeId]; !ok {
					themeMap[theme.ThemeId] = theme
					IconThemes.Put(theme)
				}
			}
		}
	}
	return themeMap
}

func readTheme(indexThemeFilePath string) (*IconTheme, bool) {
	// id of a theme is the name of the directory in which its index.theme file resides
	var themeId = filepath.Base(filepath.Dir(indexThemeFilePath))

	if themeId == "." || themeId == ".." || themeId == "/" {
		log.Warn("Could not figure theme id from path:", indexThemeFilePath)
	}

	iniFile, err := xdg.ReadIniFile(indexThemeFilePath)
	if err != nil {
		log.Warn("Error reading theme:", err)
		return nil, false
	}

	if len(iniFile) < 1 || iniFile[0].Name != "Icon Theme" {
		log.Warn("Error reading %s , expected 'Icon Theme' at start", indexThemeFilePath)
		return nil, false
	}

	themeGroup := iniFile[0]

	theme := IconTheme{}
	theme.ThemeId = themeId
	theme.self = "/icontheme/" + themeId
	theme.Name = themeGroup.Entries["Name"]
	theme.Comment = themeGroup.Entries["Comment"]
	theme.Inherits = slice.Split(themeGroup.Entries["Inherits"], ",")
	theme.Dirs = make([]IconDir, 0, 50)
	var addedDirs = make(map[string]bool)
	directories := slice.Split(themeGroup.Entries["Directories"], ",")
	if len(directories) == 0 {
		log.Warn("Ignoring theme ", theme.ThemeId, " - no directories")
		return nil, false
	}
	for _, iniGroup := range iniFile[1:] {

		if !slice.Contains(directories, iniGroup.Name) {
			//fmt.Fprintln(os.Stderr, iniGroup.Name, " not found in Directories")
			continue
		}

		if addedDirs[iniGroup.Name] {
			log.Warn(iniGroup.Name, "encountered more than once")
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

		theme.Dirs = append(theme.Dirs, IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entries["Context"]})
		addedDirs[iniGroup.Name] = true
	}

	if themeId == "hicolor" {
		// We have icons from other sources than the various icon directories.. x11, statusnotifications, notifications. In order to serve those in
		// a way similar to ordinary icons - we add them either as not-themed icons or as icons in the hicolor theme.
		// Here we create an extra set of directories for hicolor - context 'converted' - where we put those
		var lowerBound uint32 = 0
		for _, upperBound := range []uint32{16, 22, 24, 32, 36, 48, 64, 72, 96, 128, 192, 256, 512, 1024, 2048} {
			var path = fmt.Sprintf("converted/%d", upperBound)
			if addedDirs[path] {
				panic(path + " in use")
			}
			theme.Dirs = append(theme.Dirs, IconDir{path, lowerBound, upperBound, "converted"})
			lowerBound = upperBound + 1
		}
	}

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

var IconThemes = resource.MakeCollection[*IconTheme]()
