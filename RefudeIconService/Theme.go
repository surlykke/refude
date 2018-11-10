// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"log"
	"os"
	"strconv"
	"strings"
)

type Theme struct {
	Id          string
	Name        string
	Comment     string
	Context     string
	Inherits    []string
	SearchOrder []string
	IconDirs    []IconDir
	Icons       map[string]Icon
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}


var searchDirectories []string = getSearchDirectories()



func readIndexTheme(themeId string, indexThemeFilePath string) (Theme, error) {
	fmt.Println("readIndexTheme, path:", indexThemeFilePath)
	iniFile, err := xdg.ReadIniFile(indexThemeFilePath)
	if err != nil {
		log.Println("Error reading theme:", err)
		return Theme{}, err
	}

	if len(iniFile) < 1 || iniFile[0].Name != "Icon Theme" {
		return Theme{}, fmt.Errorf("Error reading %s , expected 'Icon Theme' at start", indexThemeFilePath)
	}

	themeGroup := iniFile[0]

	theme := Theme{}
	theme.Id = themeId
	theme.Name = themeGroup.Entries["Name"]
	theme.Comment = themeGroup.Entries["Comment"]
	theme.Inherits = slice.Split(themeGroup.Entries["Inherits"], ",")
	theme.IconDirs = []IconDir{}
	directories := slice.Split(themeGroup.Entries["Directories"], ",")
	for _, iniGroup := range iniFile[1:] {

		if !slice.Contains(directories, iniGroup.Name) {
			fmt.Fprintln(os.Stderr, iniGroup.Name, " not found in Directories")
			continue
		}

		size, sizeGiven := readUint32(iniGroup.Entries["Size"])
		if !sizeGiven {
			fmt.Fprintln(os.Stderr, "Skipping ", iniGroup.Name, " - no size given")
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
			fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name,
				                    ", type must be given as 'Fixed', 'Scalable' or 'Threshold', was: ", sizeType)
			continue
		}

		theme.IconDirs = append(theme.IconDirs, IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entries["Context"]})
	}

	theme.Icons = make(map[string]Icon)

	return theme, nil
}

func getAncestors(themeId string, visited []string, themeMap map[string]Theme) []string {
	ancestors := make([]string, 0)
	if themeId != "hicolor" && !slice.Contains(visited, themeId) {
		slice.AppendIfNotThere(visited, themeId)
		if theme, ok := themeMap[themeId]; ok {
			ancestors = append(ancestors, themeId)
			for _, parentId := range theme.Inherits {
				ancestors = append(ancestors, getAncestors(parentId, visited, themeMap)...)
			}
		}
	}

	return ancestors
}

func reverse(strings []string) []string {
	res := make([]string, len(strings))
	for i := range strings {
		res[len(strings)-1-i] = strings[i]
	}
	return res
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
