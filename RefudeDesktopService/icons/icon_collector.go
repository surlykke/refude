// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"crypto/sha1"
	"fmt"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var iconCollection = MakeIconCollection()
var dirIsAdded = make(map[string]bool)

func addBaseDir(dir string) {
	if ! dirIsAdded[dir] {
		dirIsAdded[dir] = true
		// First look for theme definitions
		if indexThemeFilePaths, err := filepath.Glob(dir + "/" + "*" + "/index.theme"); err != nil {
			log.Println("Error reading", dir, "-", err)
		} else {
			for _, indexThemeFilePath := range indexThemeFilePaths {
				discoveredThemes <- struct {
					basedir  string;
					themedir string
				}{dir, filepath.Base(filepath.Dir(indexThemeFilePath))}
			}
		}

		// Then collect icons

		fmt.Println("Collect from", dir)
		_ = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				log.Println("Error descending into", path)
				return err
			} else {
				if !info.IsDir() &&
					(strings.HasSuffix(info.Name(), ".png") ||
						strings.HasSuffix(info.Name(), ".xpm") ||
						strings.HasSuffix(info.Name(), ".svg")) {
					var icon = &Icon{Name: info.Name()[0 : len(info.Name())-4], Path: path, Type: info.Name()[len(info.Name())-3:]}
					if len(path)-len(info.Name())-2 > len(dir) {
						var relIconDir = path[len(dir)+1 : len(path)-len(info.Name())-1]
						if slashPos := strings.Index(relIconDir, "/"); slashPos > 0 {
							icon.Theme = relIconDir[0:slashPos]
							icon.themeSubDir = relIconDir[slashPos+1:]
						} else {
							icon.Theme = relIconDir
						}
					}

					//fmt.Printf("%s below %s - Name: '%s', Type: '%s', Theme: '%s', themeSubDir: '%s'\n", path, dir, icon.Name, icon.Type, icon.Theme, icon.themeSubDir)

					icon.AbstractResource = resource.MakeAbstractResource(resource.Standardize("/icon/"+url.PathEscape(path)), "application/json")

					discoveredIcons <- icon

				}
				return nil
			}
		})

	}
}

var discoveredThemes = make(chan struct {
	basedir  string;
	themedir string
})

var discoveredIcons = make(chan *Icon)

func consumeThemes() {
	var haveSeenThemeDir = make(map[string]bool)
	for discoveredTheme := range discoveredThemes {
		if !haveSeenThemeDir[discoveredTheme.themedir] {
			if theme, err := readIndexTheme(discoveredTheme.themedir, discoveredTheme.basedir+"/"+discoveredTheme.themedir+"/index.theme"); err == nil {
				iconCollection.addTheme(theme)
			}
		}
	}
}

func consumeIcons() {
	for discoveredIcon := range discoveredIcons {
		iconCollection.addIcon(discoveredIcon)
	}
}

func readIndexTheme(themeId string, indexThemeFilePath string) (*Theme, error) {
	iniFile, err := xdg.ReadIniFile(indexThemeFilePath)
	if err != nil {
		log.Println("Error reading theme:", err)
		return nil, err
	}

	if len(iniFile) < 1 || iniFile[0].Name != "Icon Theme" {
		return nil, fmt.Errorf("Error reading %s , expected 'Icon Theme' at start", indexThemeFilePath)
	}

	themeGroup := iniFile[0]

	theme := &Theme{}
	theme.Id = themeId
	theme.Name = themeGroup.Entries["Name"]
	theme.Comment = themeGroup.Entries["Comment"]
	theme.Inherits = slice.Remove(slice.Split(themeGroup.Entries["Inherits"], ","), "hicolor"); // hicolor is always inherited, so no reason to list.
	theme.iconDirs = make(map[string]IconDir)
	directories := slice.Split(themeGroup.Entries["Directories"], ",")
	if len(directories) == 0 {
		return nil, fmt.Errorf("Ignoring theme", theme.Id, "- no directories")
	}
	fmt.Println("Theme", theme.Name, "directories:", directories)
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
			_, _ = fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name, ", type must be given as 'Fixed', 'Scalable' or 'Threshold', was: ", sizeType)
			continue
		}

		theme.iconDirs[iniGroup.Name] = IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entries["Context"]}
	}

	theme.icons = make(map[string][]*Icon)

	theme.AbstractResource = resource.MakeAbstractResource(resource.Standardizef("/icontheme/%s", theme.Id), "application/json")
	return theme, nil
}

func getPathToConverted(pathToXpm string) (string, error) {
	if xpmBytes, err := ioutil.ReadFile(pathToXpm); err != nil {
		return "", err
	} else {
		pngPath := fmt.Sprintf("%s/%x.png", xdg.RefudeConvertedIconsDir, sha1.Sum(xpmBytes))
		if _, err := os.Stat(pngPath); os.IsNotExist(err) {
			if pngBytes, err := image.Xpm2png(xpmBytes); err != nil {
				return "", err
			} else if err = ioutil.WriteFile(pngPath, pngBytes, 0700); err != nil {
				return "", err
			}
		} else if err != nil {
			return "", err
		}

		return pngPath, nil
	}
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
