// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"sync"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/slice"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type nameAndSize struct {
	name string
	size uint32
}

var (
	lock                  sync.Mutex
	basedirs              = make([]string, 0, 10)
	themeMap              = make(map[string]*IconTheme)
	themeSearchList       = []string{} // id's of themes to search. First, defaultTheme, if given, then those directly or indirectly inherited, ending with hicolor if installed
	iconPathCache         = make(map[nameAndSize]string)
	addedFiles            = make(map[string]struct{}) // We only serve icons by absolute path if found here
	addedSessionIcons     = make(map[string]struct{})
	refudeSessionIconsDir string
)

func init() {
	refudeSessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"
	if err := os.MkdirAll(refudeSessionIconsDir, 0700); err != nil {
		panic(err)
	}

	determineBasedirs()
	themeMap = readThemes()
	determineDefaultIconTheme()

	addInheritedThemesToThemeList()

	if _, ok := themeMap["hicolor"]; ok {
		// We lay out a hicolor directory structure in sessionsdir
		var sessionHicolorPath = refudeSessionIconsDir + "/hicolor/"
		for _, dir := range themeMap["hicolor"].Dirs {
			os.MkdirAll(sessionHicolorPath+dir.Path, 0700)
		}
	}

	var iconThemeList = make([]*IconTheme, 0, len(themeMap))
	for _, iconTheme := range themeMap {
		iconThemeList = append(iconThemeList, iconTheme)
	}
	IconThemes.ReplaceWith(iconThemeList)
}

func determineBasedirs() {
	var tmp = make([]string, 0, 10)
	tmp = append(tmp, xdg.Home+"/.icons")
	tmp = append(tmp, xdg.DataHome+"/icons") // Weirdly icontheme specification does not mention this, which I consider to be an error
	for _, dataDir := range xdg.DataDirs {
		tmp = append(tmp, dataDir+"/icons")
	}
	tmp = append(tmp, "/usr/share/pixmaps")
	for _, dir := range tmp {
		if existsAndIsDir(dir) {
			basedirs = append(basedirs, dir)
		}
	}
	basedirs = append(basedirs, refudeSessionIconsDir)
}

/**
 *Finds, if possible, the default theme, and places it first in themeList
 */
func determineDefaultIconTheme() {
	var (
		defaultThemeName = ""
		ok               bool
		filesToLookAt    = []string{
			xdg.ConfigHome + "/gtk-4.0/settings.ini",
			"/etc/gtk-4.0/settings.ini",
			xdg.ConfigHome + "/gtk-3.0/settings.ini",
			"/etc/gtk-3.0/settings.ini",
			xdg.ConfigHome + "/gtk-2.0/settings.ini",
			"/etc/gtk-2.0/settings.ini",
			xdg.Home + "/.gtkrc-2.0",
			"/etc/gtk-2.0/gtkrc",
		}

		iconThemeDefPattern = regexp.MustCompile("gtk-icon-theme-name=(\\S+)")
	)

	if defaultThemeName, ok = os.LookupEnv("REFUDE_ICON_THEME"); ok {
		log.Info("default icon theme taken from env REFUDE_ICON_THEME", defaultThemeName)
	} else {
		for _, fileToLookAt := range filesToLookAt {
			if bytes, err := ioutil.ReadFile(fileToLookAt); err == nil {
				if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
					defaultThemeName = matches[1]
					log.Info("default icon theme taken from", fileToLookAt, defaultThemeName)
					break
				}
			}
		}
	}

	if defaultThemeName != "" {
		for themeId, theme := range themeMap {
			if theme.Title == defaultThemeName {
				themeSearchList = []string{themeId}
				return
			}
		}
	}
}

func addInheritedThemesToThemeList() {
	for i := 0; i < len(themeSearchList); i++ {
		theme := themeMap[themeSearchList[i]]
		for _, inheritedId := range theme.Inherits {
			if _, ok := themeMap[inheritedId]; ok {
				themeSearchList = slice.AppendIfNotThere(themeSearchList, inheritedId)
			}
		}
	}

	// Move hicolor to last
	if _, ok := themeMap["hicolor"]; ok {
		themeSearchList = slice.Remove(themeSearchList, "hicolor")
		themeSearchList = append(themeSearchList, "hicolor")
	}
}

func AddX11Icon(data []uint32) (string, error) {
	var iconName = image.X11IconHashName(data)

	lock.Lock()
	defer lock.Unlock()
	if _, ok := addedSessionIcons[iconName]; !ok {
		addedSessionIcons[iconName] = struct{}{}

		if pngList, err := image.X11IconToPngs(data); err != nil {
			log.Warn("Error converting:", err)
			return "", err
		} else {
			for _, sizedPng := range pngList {
				if sizedPng.Width != sizedPng.Height {
					log.Warn("Ignore image", sizedPng.Width, "x", sizedPng.Height, ", not square")
				} else {
					writeSessionHicolorIcon(iconName, sizedPng.Height, sizedPng.Data)
				}
			}
		}

	} else {
	}

	return iconName, nil
}

func AddARGBIcon(argbIcon image.ARGBIcon) string {
	var iconName = image.ARGBIconHashName(argbIcon)
	lock.Lock()
	defer lock.Unlock()

	if _, ok := addedSessionIcons[iconName]; !ok {
		for _, pixMap := range argbIcon.Images {
			if pixMap.Width == pixMap.Height { // else ignore
				if png, err := pixMap.AsPng(); err != nil {
					log.Warn("Unable to convert image", err)
				} else {
					writeSessionHicolorIcon(iconName, pixMap.Height, png)
				}
			}
		}

	}
	return iconName

}

func AddFileIcon(filePath string) {
	lock.Lock()
	defer lock.Unlock()
	addedFiles[filePath] = struct{}{}
}

func AddRawImageIcon(imageData image.ImageData) string {
	iconName := image.ImageDataHashName(imageData)
	lock.Lock()
	defer lock.Unlock()
	if _, ok := addedSessionIcons[iconName]; !ok {
		if png, err := imageData.AsPng(); err != nil {
			log.Warn("Error converting image", err)
			return ""
		} else {
			writeSessionOtherIcon(iconName, png)
		}
	}
	return iconName
}

func writeSessionHicolorIcon(iconName string, size uint32, data []byte) {
	for _, dir := range themeMap["hicolor"].Dirs {
		if dir.Context == "converted" && dir.MinSize <= size && dir.MaxSize >= size {
			var path = fmt.Sprintf("%s/hicolor/%s/%s.png", refudeSessionIconsDir, dir.Path, iconName)
			if err := ioutil.WriteFile(path, data, 0700); err != nil {
				log.Warn("Problem writing", path, err)
			}
			return
		}
	}

	log.Warn("Found no suitable converted dir for", iconName, "of size", size)
}

func writeSessionOtherIcon(iconName string, data []byte) {
	var path = fmt.Sprintf("%s/%s.png", refudeSessionIconsDir, iconName)
	if err := ioutil.WriteFile(path, data, 0700); err != nil {
		log.Warn("Problem writing", path, err)
	}
}

func AddBasedir(basedir string) {
	lock.Lock()
	defer lock.Unlock()
	basedirs = slice.AppendIfNotThere(basedirs, basedir)
}
