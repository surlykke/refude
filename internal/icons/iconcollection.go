// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package icons

import (
	"errors"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/xdg"
)

func init() {
}

type Descriptor struct {
	osPath           string
	themeId          string
	minSize, maxSize uint32
}

func collectIcons() {
	if defaultTheme := determineDefaultTheme(); defaultTheme != nil {
		log.Print("No defaultTheme determinded. Cannot serve icons")
	} else if themeList, err := themeSearchOrder(defaultTheme); err != nil {
		log.Print("No theme search order determined. Cannot serve icons")
	} else {
		for desc := range iconPaths(themeList, "svg") {
			addIconPath(desc.osPath, desc.themeId, desc.minSize, desc.maxSize)
		}
	}
}

func addIconPath(osPath, themeId string, minSize, maxSize uint32) {
	if realPath, err := filepath.EvalSymlinks(osPath); err != nil {
		log.Println("Error adding iconPath", osPath, err)
	} else if iconName := path.Base(realPath[:len(realPath)-4]); iconName == "" {
		log.Println("Weird iconName:", osPath)
	} else {
		iconPtr, ok := resource.Get[*Icon](iconPrefix + iconName)
		var icon Icon
		if ok {
			icon = *iconPtr
		} else {
			icon = Icon{Base: resource.MakeBase(iconPrefix+iconName, iconName, "", ""), Name: iconName, Theme: themeId}
		}
		if !icon.HasPath(osPath) {
			icon.Imgs = append(icon.Imgs, Img{OsPath: osPath, MinSize: minSize, MaxSize: maxSize})
		}
		resource.Put(&icon)
	}
}

func iconPaths(themeList []*IconTheme, fileType string) func(func(Descriptor) bool) {
	return func(yield func(Descriptor) bool) {
		for _, theme := range themeList {
			var themeId = theme.Id
			for _, basedir := range xdg.IconBasedirs {
				for _, themeDir := range theme.Dirs {
					var paths, _ = filepath.Glob(basedir + "/" + themeId + "/" + themeDir.Path + "/*." + fileType)
					for _, path := range paths {
						if !yield(Descriptor{osPath: path, themeId: theme.Id, minSize: themeDir.MinSize, maxSize: themeDir.MaxSize}) {
							return
						}
					}
				}
			}

			dirsWithNonThemedIcons := append(xdg.IconBasedirs, xdg.PixmapDir)
			for _, dir := range dirsWithNonThemedIcons {
				var paths, _ = filepath.Glob(dir + "/*")
				for _, path := range paths {
					if !yield(Descriptor{osPath: path, minSize: 1, maxSize: 1}) {
						return
					}
				}
			}

		}
	}
}

func determineDefaultTheme() *IconTheme {
	var iconThemeDefPattern = regexp.MustCompile(`gtk-icon-theme-name=(\S+)`)

	if themeId, ok := os.LookupEnv("REFUDE_ICON_THEME"); ok {
		if theme, ok := resource.Get[*IconTheme](themePathPrefix + themeId); ok {
			return theme
		}
	}

	var iniFilesToLookAt = []string{
		xdg.ConfigHome + "/gtk-4.0/settings.ini", "/etc/gtk-4.0/settings.ini",
		xdg.ConfigHome + "/gtk-3.0/settings.ini", "/etc/gtk-3.0/settings.ini",
		xdg.ConfigHome + "/gtk-2.0/settings.ini", "/etc/gtk-2.0/settings.ini",
		xdg.Home + "/.gtkrc-2.0", "/etc/gtk-2.0/gtkrc"}

	for _, iniFile := range iniFilesToLookAt {
		if bytes, err := os.ReadFile(iniFile); err == nil {
			if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
				if theme, ok := resource.Get[*IconTheme](matches[1]); ok {
					return theme
				}
			}
		}
	}
	return nil
}

func themeSearchOrder(theme *IconTheme) ([]*IconTheme, error) {

	var hicolorTheme, ok = resource.Get[*IconTheme](themePathPrefix + "hicolor")
	if !ok {
		return nil, errors.New("No hicolortheme found")
	}
	var order = make([]*IconTheme, 1, 10)
	order[0] = theme
	var pos = 0
	for pos < len(order) {
		for _, inheritedId := range order[pos].Inherits {
			if inheritedTheme, ok := resource.Get[*IconTheme](themePathPrefix + inheritedId); ok && inheritedTheme != hicolorTheme {
				if !slices.Contains(order, inheritedTheme) {
					order = append(order, inheritedTheme)
				}
			}
		}
		pos++
	}

	return append(order, hicolorTheme), nil
}
