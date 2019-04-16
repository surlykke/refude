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
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var refudeConvertedIconsDir string
var refudeSessionIconsDir string

var addedBaseDirs map[string]bool
var unplacedIcons = []*Icon{}

func init() {
	refudeConvertedIconsDir = xdg.RuntimeDir + "/org.refude.converted-icons"
	refudeSessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"

	if err := os.MkdirAll(refudeConvertedIconsDir, 0700); err != nil {
		panic(err)
	}

	if err := os.MkdirAll(refudeSessionIconsDir, 0700); err != nil {
		panic(err)
	}
	themes = make(map[string]*Theme)
	themeIcons = make(map[string]map[string][]*Icon)
	otherIcons = make(map[string]*Icon)
	iconsByPath = make(map[resource.StandardizedPath]*Icon)
	addedBaseDirs = make(map[string]bool)
}

func monitorBasedirSink() {

	for baseDir := range basedirSink {
		addBaseDir(baseDir)
	}

}

func monitorIconSink() {
	for icon := range iconSink {
		addARGBIcon(icon)
	}
}

func addBaseDir(baseDir string) {
	baseDir = path.Clean(baseDir)
	if addedBaseDirs[baseDir] {
		return
	}

	// First look for theme definitions
	if indexThemeFilePaths, err := filepath.Glob(baseDir + "/" + "*" + "/index.theme"); err != nil {
		log.Println("Error globbing", baseDir, "-", err)
	} else {
		for _, indexThemeFilePath := range indexThemeFilePaths {
			var indexThemeFileRelativePath = indexThemeFilePath[len(baseDir)+1:]
			var themeId = filepath.Base(filepath.Dir(indexThemeFileRelativePath))
			fmt.Println("look at", themeId, "in", baseDir)
			if themeId != "" && getTheme(themeId) == nil {
				if theme, err := readIndexTheme(themeId, indexThemeFilePath); err != nil {
					fmt.Println("Error reading", indexThemeFilePath, err)
				} else {
					addTheme(theme)

					var unplacedIconsCount = 0
					for i := 0; i < len(unplacedIcons); i++ {
						if !placeIcon(themeId, unplacedIcons[i]) {
							unplacedIcons[unplacedIconsCount] = unplacedIcons[i]
							unplacedIconsCount++
						}
					}
					unplacedIcons = unplacedIcons[0:unplacedIconsCount]
				}
			}
		}
	}

	_ = filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Println("Error descending into", path)
			return err
		} else {
			if !info.IsDir() &&
				(strings.HasSuffix(info.Name(), ".png") ||
					strings.HasSuffix(info.Name(), ".xpm") ||
					strings.HasSuffix(info.Name(), ".svg")) {

				var icon = &Icon{Name: info.Name()[0 : len(info.Name())-4], Path: path, Type: info.Name()[len(info.Name())-3:]}
				if len(path)-len(info.Name())-2 > len(baseDir) {
					var relIconDir = path[len(baseDir)+1 : len(path)-len(info.Name())-1]
					if slashPos := strings.Index(relIconDir, "/"); slashPos > 0 {
						icon.Theme = relIconDir[0:slashPos]
						icon.themeSubdir = relIconDir[slashPos+1:]
					} else {
						icon.Theme = relIconDir
					}
				}

				icon.AbstractResource = resource.MakeAbstractResource(resource.Standardize("/icon/"+url.PathEscape(path)), "application/json")
				if icon.Theme != "" {
					if !placeIcon(icon.Theme, icon) {
						unplacedIcons = append(unplacedIcons, icon)
					}
				} else {
					addOtherIcon(icon)
				}
			}
			return nil
		}
	})
}

func placeIcon(themeId string, icon *Icon) bool {
	if theme := getTheme(themeId); theme != nil {
		if dir, ok := theme.Dirs[icon.themeSubdir]; ok {
			icon.Context = dir.Context
			icon.MinSize = dir.MinSize
			icon.MaxSize = dir.MaxSize

			addThemeIcon(theme.Id, icon)
			return true
		}

	}
	return false
}

func addARGBIcon(argbIcon image.ARGBIcon) {
	fmt.Println("add icon", argbIcon.Name)

	fmt.Println("Adding icon", argbIcon.Name, "to hicolorTheme")

	if !haveThemeIcon("hicolor", argbIcon.Name) {

	}

	var hicolorIconMap = themeIcons["hicolor"]

	if _, ok := hicolorIconMap[argbIcon.Name]; !ok {
		for _, pixMap := range argbIcon.Images {
			if pixMap.Width != pixMap.Height {
			} else {
				var path = fmt.Sprintf("%s/%d/%s.png", refudeSessionIconsDir, pixMap.Width, argbIcon.Name)
				var icon = &Icon{
					Name:    argbIcon.Name,
					Theme:   "hicolor",
					Context: "",
					Type:    "png",
					MinSize: pixMap.Width,
					MaxSize: pixMap.Width,
					Path:    path,
				}
				addThemeIcon("hicolor", icon)
				go savePng(path, pixMap)
			}
		}
	}
}

func savePng(path string, pixMap image.ARGBImage) {
	if png, err := pixMap.AsPng(); err != nil {
		log.Println("Error converting pixmap to png:", err)
	} else {
		var lastSlashPos = strings.LastIndex(path, "/")
		var dir = path[0:lastSlashPos]
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Println("Unable to create", dir, err)
		} else if err = ioutil.WriteFile(path, png, 0700); err != nil {
			log.Println("Unable to write file", err)
		}
	}

}

func findMatchingDir(theme *Theme, size uint32, context string) (IconDir, bool) {
	for _, dir := range theme.Dirs {
		if dir.MinSize <= size && dir.MaxSize >= size && dir.Context == context {
			return dir, true
		}
	}

	return IconDir{}, false
}

func readIndexTheme(themeId string, indexThemeFilePath string) (*Theme, error) {
	iniFile, err := xdg.ReadIniFile(indexThemeFilePath)
	if err != nil {
		//log.Println("Error reading theme:", err)
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
	theme.Inherits = slice.Remove(slice.Split(themeGroup.Entries["Inherits"], ","), "hicolor") // hicolor is always inherited, so no reason to list.
	theme.Dirs = make(map[string]IconDir)
	directories := slice.Split(themeGroup.Entries["Directories"], ",")
	if len(directories) == 0 {
		return nil, fmt.Errorf("Ignoring theme %s - no directories", theme.Id)
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

	theme.AbstractResource = resource.MakeAbstractResource(resource.Standardizef("/icontheme/%s", theme.Id), "application/json")
	return theme, nil
}

func getPathToConverted(pathToXpm string) (string, error) {
	if xpmBytes, err := ioutil.ReadFile(pathToXpm); err != nil {
		return "", err
	} else {
		pngPath := fmt.Sprintf("%s/%x.png", refudeConvertedIconsDir, sha1.Sum(xpmBytes))
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
