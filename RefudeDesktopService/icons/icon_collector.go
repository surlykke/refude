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
	"github.com/surlykke/RefudeServices/lib/fs"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func Run(collection *IconCollection) {
	go collectAndMonitorThemeIcons(collection)
	go collectAndMonitorOtherIcons(collection)
}

func collectAndMonitorThemeIcons(collection *IconCollection) {
	collection.SetThemes(collectThemes()) // TODO Should loop, monitoring theme directories, and recollect on change
}

func collectThemes() map[string]*Theme {
	var iconDirs = []string{xdg.Home + "/.icons", xdg.DataHome + "/icons"}
	for i := len(xdg.DataDirs) - 1; i >= 0; i-- {
		iconDirs = append(iconDirs, xdg.DataDirs[i]+"/icons")
	}

	themes := make(map[string]*Theme)

	for _, searcDir := range iconDirs {
		indexThemeFilePaths, err := filepath.Glob(searcDir + "/" + "*" + "/index.theme")
		if err != nil {
			panic(err)
		}

		for _, indexThemeFilePath := range indexThemeFilePaths {
			fmt.Println("Looking at: ", indexThemeFilePath)
			themeId := filepath.Base(filepath.Dir(indexThemeFilePath))
			if _, ok := themes[themeId]; !ok {
				if theme, err := readIndexTheme(themeId, indexThemeFilePath); err == nil {
					themes[themeId] = theme
				} else {
					log.Println("Error reading index.theme: ", err)
				}

			}
		}
	}

	for themeId, theme := range themes {
		if themeId != "hicolor" {
			theme.SearchOrder = []string{themeId}
			for i := 0; i < len(theme.SearchOrder); i++ {
				var themeToSearch= themes[theme.SearchOrder[i]]
				for _, parentId := range themeToSearch.Inherits {
					if parentTheme, ok := themes[parentId]; ok && parentTheme.Id != "hicolor" {
						theme.SearchOrder = slice.AppendIfNotThere(theme.SearchOrder, parentId)
					}
				}
			}
		}

		theme.SearchOrder = append(theme.SearchOrder, "hicolor")

		for _, searchDir := range iconDirs {
			for _, iconDir := range theme.iconDirs {
				for _, icon := range collectIcons(searchDir + "/" + themeId + "/" + iconDir.Path) {
					icon.Theme = themeId
					icon.Context = iconDir.Context
					icon.MinSize = iconDir.MinSize
					icon.MaxSize = iconDir.MaxSize
					theme.icons[icon.Name] = append(theme.icons[icon.Name], icon)
				}

			}
		}

	}

	return themes
}

func collectAndMonitorOtherIcons(collection *IconCollection) {
	var dirs = []string{"/usr/share/pixmaps", xdg.RefudeSessionIconsDir, xdg.RefudeConvertedIconsDir}
	if watcher, err := fs.MakeWatcher(dirs...); err != nil {
		log.Println("Unable to watch", dirs, err)
	} else {
		var icons = make(map[string]*Icon)
		for _, dir := range []string{"/usr/share/pixmaps", xdg.RefudeConvertedIconsDir, xdg.RefudeSessionIconsDir} {
			for _, icon := range collectIcons(dir) {
				icons[icon.Name] = icon
			}
		}
		collection.SetOtherIcons(icons)

		fs.Wait(watcher)

		time.Sleep(time.Second * 2)

		go collectAndMonitorOtherIcons(collection)
	}
}

func readIndexTheme(themeId string, indexThemeFilePath string) (*Theme, error) {
	fmt.Println("readIndexTheme, path:", indexThemeFilePath)
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
	theme.Inherits = slice.Split(themeGroup.Entries["Inherits"], ",")
	theme.iconDirs = []IconDir{}
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
			_, _ = fmt.Fprintln(os.Stderr, "Error in ", theme.Name, ", ", iniGroup.Name, ", type must be given as 'Fixed', 'Scalable' or 'Threshold', was: ", sizeType)
			continue
		}

		theme.iconDirs = append(theme.iconDirs, IconDir{iniGroup.Name, minSize, maxSize, iniGroup.Entries["Context"]})
	}

	theme.icons = make(map[string][]*Icon)

	return theme, nil
}

// Returned icons only partially instantiated, caller must add Context, MinSize, MaxSize
func collectIcons(dir string) []*Icon {
	fmt.Print("Collecting in: ", dir)
	var icons = make([]*Icon, 0, 100)
	for _, ending := range []string{"png", "svg", "xpm"} {
		paths, err := filepath.Glob(dir + "/*." + ending)
		if err != nil {
			panic(err)
		}

		for _, path := range paths {
			var iconName= filepath.Base(path[0 : len(path)-4])

			if ending == "xpm" {
				if tmp, err := getPathToConverted(path); err != nil {
					log.Println("Unable to convert", path, ":", err)
					continue
				} else {
					fmt.Println("Converting", path, "to", tmp)
					path = tmp
				}
			}
			var icon= &Icon{Name: iconName, Type: ending, img: img{Path: path}}
			icon.AbstractResource = resource.MakeAbstractResource(resource.Standardize("/icon" + path), "application/json")
			icons = append(icons, icon)
		}
	}
	fmt.Println(" found", len(icons))
	return icons
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
