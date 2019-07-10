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
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

/** Urls
 * /icons   list of icons for curren default theme
 * /allicons list of all icons for all installed themes
 * /icon/<name> json representation of icon <name>
 */

/**
 * Example naming of directory variables in this code:
 *
 * /usr/share/icons/oxygen/base/32x32/actions/
 * |-datadir-|
 * |----basedir----|
 * |--------themedir------|
 * |----------------themesubdir--------------|
 */

var refudeConvertedIconsDir string
var refudeSessionIconsDir string

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

type IconCollection struct {
	sync.Mutex
	basedirs   []string
	themes     map[string]*Theme
	themeIcons map[string]map[string]*Icon
	otherIcons map[string]*Icon
}

func (ic *IconCollection) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	if strings.HasPrefix(r.URL.Path, "/theme/") {
		if theme := ic.findTheme(r.URL.Path[7:]); theme != nil {
			theme.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if strings.HasPrefix(r.URL.Path, "/icon/") {
		if icon := ic.findIcon("oxygen" /*FIXME*/, r.URL.Path[6:]); icon != nil {
			icon.ServeHTTP(w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		return false
	}
	return true
}

func MakeIconCollection() *IconCollection {
	var ic = IconCollection{
		basedirs:   []string{xdg.Home + "/.icons", xdg.Home + "/.local/share/icons"},
		themes:     make(map[string]*Theme),
		themeIcons: make(map[string]map[string]*Icon),
		otherIcons: make(map[string]*Icon),
	}

	for _, datadir := range xdg.DataDirs {
		ic.basedirs = append(ic.basedirs, datadir+"/icons")
	}
	ic.basedirs = append(ic.basedirs, "/usr/share/pixmaps")
	return &ic
}

func (ic *IconCollection) collect() {
	ic.Lock()
	defer ic.Unlock()

	for _, basedir := range ic.basedirs {
		ic.collectThemes(basedir)
	}

	for _, basedir := range ic.basedirs {
		ic.collectThemeicons(basedir)
		ic.collectOtherIcons(basedir)
	}
}

func (ic *IconCollection) addBasedir(basedir string) {
	ic.Lock()
	defer ic.Unlock()

	if !slice.Contains(ic.basedirs, basedir) {
		ic.basedirs = append(ic.basedirs, basedir)
		ic.collectThemeicons(basedir)
		ic.collectOtherIcons(basedir)
	}
}

//
// Adds the icon to hicolor
//
func (ic *IconCollection) addARGBIcon(argbIcon image.ARGBIcon) {
	if _, ok := ic.themeIcons["hicolor"][argbIcon.Name]; ok {
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
		ic.Lock()
		defer ic.Unlock()
		ic.themeIcons["hicolor"][icon.Name] = &icon
	}
}

// Caller holds lock
func (ic *IconCollection) collectThemes(basedir string) {
	fmt.Println("Collecting themes in", basedir)
	basesubdirs, err := getVisibleSubdirs(basedir)
	fmt.Println("basesubdirs:", basesubdirs)
	if err != nil {
		fmt.Println("err:", err)
		return
	}
	for _, basesubdir := range basesubdirs {
		var themedir = basedir + "/" + basesubdir
		var themeName = filepath.Base(themedir)
		fmt.Println("themeName:", themeName)
		if _, ok := ic.themes[themeName]; ok {
			continue
		}

		var themeIndexPath = themedir + "/index.theme"
		_, err := os.Stat(themeIndexPath)
		if err != nil {
			fmt.Println("err:", err)
			if !os.IsNotExist(err) {
				fmt.Println("Error accessing", themeIndexPath, ":", err)
			}
			continue
		}

		theme, err := readIndexTheme(themeName, themeIndexPath)
		if err != nil {
			fmt.Println("Error reading", themeIndexPath, ":", err)
			continue
		}
		fmt.Println("Adding theme", themeName)
		ic.themes[themeName] = theme
		ic.themeIcons[themeName] = make(map[string]*Icon)
	}
}

// Caller holds lock
func (ic *IconCollection) collectThemeicons(icondir string) {
	for themename, theme := range ic.themes {
		var themedir = icondir + "/" + themename
		if !dirExists(themedir) {
			continue
		}
		for _, dir := range theme.Dirs {
			var themesubdir = themedir + "/" + dir.Path
			if !dirExists(themesubdir) {
				continue
			}

			iconFileNames, err := getIcons(themesubdir)
			if err != nil {
				fmt.Println("Error reading icons in", themesubdir, ":", err)
				continue
			}

			for _, iconFileName := range iconFileNames {
				var iconFilePath = themesubdir + "/" + iconFileName
				var name = iconFileName[0 : len(iconFileName)-4]
				if strings.HasSuffix(iconFilePath, ".xpm") {
					if iconFilePath, err = getPathToConverted(iconFilePath); err != nil {
						continue
					}
				}
				icon, ok := ic.themeIcons[theme.Id][name]
				if !ok {
					icon = &Icon{Name: name, Theme: themename}
					ic.themeIcons[themename][name] = icon
				}
				icon.Images = append(icon.Images, IconImage{
					Type:    iconFilePath[len(iconFilePath)-3:],
					Context: dir.Context,
					MinSize: dir.MinSize,
					MaxSize: dir.MaxSize,
					Path:    iconFilePath,
				})
			}

		}
	}
}

// Caller holds lock
func (ic *IconCollection) collectOtherIcons(basedir string) {

	if !dirExists(basedir) {
		return
	}

	iconFileNames, err := getIcons(basedir)
	if err != nil {
		log.Println("Error reading icons in", basedir, err)
	}

	for _, iconFileName := range iconFileNames {
		var iconFilePath = basedir + "/" + iconFileName
		var name = iconFileName[0 : len(iconFileName)-4]

		if strings.HasSuffix(iconFilePath, ".xpm") {
			if iconFilePath, err := getPathToConverted(iconFilePath); err != nil {
				log.Println("Problem converting", iconFilePath, err)
				continue
			}
		}

		ic.otherIcons[name] = &Icon{
			Name: name,
			Images: []IconImage{{
				Type: iconFilePath[len(iconFilePath)-3:],
				Path: iconFilePath,
			}},
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
	theme.Inherits = slice.Split(themeGroup.Entries["Inherits"], ",")
	if len(theme.Inherits) == 0 {
		theme.Inherits = []string{"hicolor"}
	}
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

	theme.Init("/icontheme/"+theme.Id, "icontheme")
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

func (ic *IconCollection) findIcon(themeId string, iconName string) *Icon {
	ic.Lock()
	defer ic.Unlock()

	var visited = make(map[string]bool)
	var toVisit = make([]string, 1, 10)
	toVisit[0] = themeId
	for i := 0; i < len(toVisit); i++ {
		var themeId = toVisit[i]
		if theme, ok := ic.themes[themeId]; ok {
			if icon, ok := ic.themeIcons[themeId][iconName]; ok {
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

	if icon, ok := ic.otherIcons[iconName]; ok {
		return icon
	}

	return nil
}

func (ic *IconCollection) findTheme(themeId string) *Theme {
	ic.Lock()
	defer ic.Unlock()

	return ic.themes[themeId]
}
