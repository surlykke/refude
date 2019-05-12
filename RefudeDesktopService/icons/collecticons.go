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
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

/** Urls
 * /icons   list of icons for curren default theme
 * /allicons list of all icons for all installed themes
 * /icon/<name> json representation of icon <name>
 */

// All use of methods in this file happens from run, so it's sequential

var refudeConvertedIconsDir string
var refudeSessionIconsDir string

var addedBaseDirs map[string]bool
var unscannedDirectories = make(map[string][]string)
var foundIconNames = make(map[string]bool)

var themes = make(map[string]*Theme)
var themeIcons = make(map[string]map[string]*Icon)
var otherIcons = make(map[string]*Icon)

func init() {
	refudeConvertedIconsDir = xdg.RuntimeDir + "/org.refude.converted-icons"
	if err := os.MkdirAll(refudeConvertedIconsDir, 0700); err != nil {
		panic(err)
	}
	refudeSessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"
	if err := os.MkdirAll(refudeSessionIconsDir, 0700); err != nil {
		panic(err)
	}

}

func addBaseDir(baseDir string) {
	baseDir = path.Clean(baseDir)
	if addedBaseDirs[baseDir] {
		return
	}

	// First, look for themes
	if baseSubdirs, err := getVisibleSubdirs(baseDir); err != nil {
		//log.Println("Error reading dirs in", baseDir, err)
	} else {
		for _, baseSubdir := range baseSubdirs {
			var themeName = baseSubdir
			var theme = themes[baseSubdir]
			var themeDirPath = baseDir + "/" + baseSubdir
			if theme == nil {
				var themeIndexPath = themeDirPath + "/index.theme"
				if _, err := os.Stat(themeIndexPath); os.IsNotExist(err) {
					unscannedDirectories[themeName] = append(unscannedDirectories[themeName], themeDirPath)
				} else if theme, err = readIndexTheme(baseSubdir, themeIndexPath); err != nil {
					//log.Println("error reading", themeIndexPath, err)
					continue
				} else {
					themes[themeName] = theme
					themeIcons[themeName] = make(map[string]*Icon)
					resourceMap.Set(theme.Self, &(*theme))

					for _, path := range unscannedDirectories[themeName] {
						collectIconsForTheme(theme, path)
					}
				}
			}

			if theme != nil {
				collectIconsForTheme(theme, themeDirPath)
			}
		}
	}

	// Then look for icons directly in base dir
	collectIconsInBaseDir(baseDir)
}

func collectIconsForTheme(theme *Theme, themeDir string) {
	for _, iconSubdir := range theme.Dirs {
		var subdirPath = themeDir + "/" + iconSubdir.Path
		if !dirExists(subdirPath) {
			continue
		}

		iconFileNames, err := getIcons(subdirPath)
		if err != nil {
			//log.Println("Error reading icons in", subdirPath, err)
			continue
		}

		for _, iconFileName := range iconFileNames {
			var iconFilePath = subdirPath + "/" + iconFileName
			var name = iconFileName[0 : len(iconFileName)-4]
			if strings.HasSuffix(iconFilePath, ".xpm") {
				if iconFilePath, err = getPathToConverted(iconFilePath); err != nil {
					continue
				}
			}
			icon, ok := themeIcons[theme.Id][name]
			if !ok {
				icon = &Icon{Name: name, Theme: theme.Id}
				themeIcons[theme.Id][name] = icon
			}
			icon.Images = append(icon.Images, IconImage{
				Type:    iconFilePath[len(iconFilePath)-3:],
				Context: iconSubdir.Context,
				MinSize: iconSubdir.MinSize,
				MaxSize: iconSubdir.MaxSize,
				Path:    iconFilePath,
			})
			foundIconNames[name] = true
		}
	}
}

func collectIconsInBaseDir(basedir string) {
	if iconFileNames, err := getIcons(basedir); err != nil {
		log.Println("Error reading icons in", basedir, err)
	} else {
		for _, iconFileName := range iconFileNames {
			var iconFilePath = basedir + "/" + iconFileName
			var name = iconFileName[0 : len(iconFileName)-4]

			if strings.HasSuffix(iconFilePath, ".xpm") {
				if iconFilePath, err := getPathToConverted(iconFilePath); err != nil {
					log.Println("Problem converting", iconFilePath, err)
					continue
				}
			}

			otherIcons[name] = &Icon{
				Name: name,
				Images: []IconImage{{
					Type: iconFilePath[len(iconFilePath)-3:],
					Path: iconFilePath,
				}},
			}

			foundIconNames[name] = true
		}
	}
}

//
// Adds the icon to hicolor
//
func addARGBIcon(argbIcon image.ARGBIcon) {
	if _, ok := themeIcons["hicolor"][argbIcon.Name]; ok {
		return
	}

	var icon = Icon{Name: argbIcon.Name, Theme: "hicolor"}

	for _, pixMap := range argbIcon.Images {
		if pixMap.Width != pixMap.Height {
		} else {
			var path = fmt.Sprintf("%s/%d/%s.png", refudeSessionIconsDir, pixMap.Width, argbIcon.Name)
			icon.Images = append(icon.Images, IconImage{
				Type:    "png",
				MinSize: pixMap.Height,
				MaxSize: pixMap.Height,
				Path:    fmt.Sprintf("%s/%d/%s.png", refudeSessionIconsDir, pixMap.Width, argbIcon.Name),
			})
			go savePng(path, pixMap)
		}
	}
	if len(icon.Images) > 0 {
		themeIcons["hicolor"][icon.Name] = &icon
		foundIconNames[icon.Name] = true
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

	theme.Self = "/icontheme/" + theme.Id
	theme.RefudeType = "icontheme"
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

func getVisibleSubdirs(dir string) ([]string, error) {
	return filterDirEntries(dir, func(info os.FileInfo) bool {
		return info.IsDir() && !strings.HasPrefix(info.Name(), ".")
	})
}

func getIcons(dir string) ([]string, error) {
	return filterDirEntries(dir, func(info os.FileInfo) bool {
		// We do not check if info represents a file. Assume that what's installed
		// in an icon directory with suffix png/svg/xpm is an imagefile or a symlink
		// to such
		return strings.HasSuffix(info.Name(), ".png") ||
			strings.HasSuffix(info.Name(), ".svg") ||
			strings.HasSuffix(info.Name(), ".xpm")
	})
}

func filterDirEntries(dir string, cond func(os.FileInfo) bool) ([]string, error) {
	if entries, err := ioutil.ReadDir(dir); err != nil {
		return nil, err
	} else {
		var result = make([]string, 0, len(entries))
		for _, entry := range entries {
			if cond(entry) {
				result = append(result, entry.Name())
			}
		}
		return result, nil
	}
}

func dirExists(dirpath string) bool {
	_, err := os.Stat(dirpath)
	return !os.IsNotExist(err)
}

func publishFoundIcons() {
	for iconName, _ := range foundIconNames {
		for themeName, _ := range themes {
			if icon := findIcon(themeName, iconName); icon != nil {
				var resolvedIcon = &(*icon)
				resolvedIcon.Self = "/icon/" + themeName + "/" + iconName
				var iconImgResource = IconImgResource{images: resolvedIcon.Images}
				resourceMap.Set(resolvedIcon.Self, resolvedIcon)
				resourceMap.Set(resolvedIcon.Self+"/img", iconImgResource)
				if themeName == "oxygen" { // FIXME
					resourceMap.Set("/icon/"+iconName, resolvedIcon)
					resourceMap.Set("/icon/"+iconName+"/img", iconImgResource)
				}
			}
		}
		delete(foundIconNames, iconName)
	}
}

func findIcon(themeId string, iconName string) *Icon {
	var visited = make(map[string]bool)
	var toVisit = make([]string, 1, 10)
	toVisit[0] = themeId
	for i := 0; i < len(toVisit); i++ {
		var themeId = toVisit[i]
		if theme, ok := themes[themeId]; ok {
			if icon, ok := themeIcons[themeId][iconName]; ok {
				return icon
			}

			visited[themeId] = true
			for _, parentId := range theme.Inherits {
				if !visited[parentId] {
					toVisit = append(toVisit, parentId)
				}
			}
		}
	}

	if themeId != "hicolor" {
		if icon, ok := themeIcons["hicolor"][iconName]; ok {
			return icon
		}
	}

	if icon, ok := otherIcons[iconName]; ok {
		return icon
	}

	return nil
}
