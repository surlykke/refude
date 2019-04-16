// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var themes map[string]*Theme
var themeLock sync.Mutex

var themeIcons map[string]map[string][]*Icon
var otherIcons map[string]*Icon
var iconsByPath map[resource.StandardizedPath]*Icon
var iconLock sync.Mutex

func GetThemes() []interface{} {
	themeLock.Lock()
	defer themeLock.Unlock()

	var themes = make([]interface{}, 0, len(themes))
	for _, theme := range themes {
		themes = append(themes, theme)
	}
	return themes
}

func getTheme(themeId string) *Theme {
	themeLock.Lock()
	defer themeLock.Unlock()
	return themes[themeId]
}

func GetTheme(path resource.StandardizedPath) *Theme {
	if !path.StartsWith("/icontheme/") {
		return nil
	} else {
		return getTheme(string(path[len("/icontheme/"):]))
	}
}

func GetIcons() []interface{} {
	iconLock.Lock()
	defer iconLock.Unlock()

	var icons = make([]interface{}, 0, 1000 /*?*/)
	for _, themeIconMap := range themeIcons {
		for _, iconList := range themeIconMap {
			for _, icon := range iconList {
				icons = append(icons, icon)
			}
		}
	}

	for _, icon := range otherIcons {
		icons = append(icons, icon)
	}

	return icons
}

// Caller ensures themeIcons[themeId] is there
func addThemeIcon(themeId string, icon *Icon) {
	iconLock.Lock()
	defer iconLock.Unlock()
	themeIcons[themeId][icon.Name] = append(themeIcons[themeId][icon.Name], icon)
}

func haveThemeIcon(themeId string, name string) bool {
	iconLock.Lock()
	defer iconLock.Unlock()

	if iconMap, ok := themeIcons[themeId]; !ok {
		return false
	} else {
		_, ok = iconMap[name]
		return ok
	}
}

func addOtherIcon(icon *Icon) {
	iconLock.Lock()
	defer iconLock.Unlock()
	otherIcons[icon.Name] = icon
}

func addTheme(theme *Theme) {
	themeLock.Lock()
	iconLock.Lock()
	defer iconLock.Unlock()
	defer themeLock.Unlock()
	themes[theme.Id] = theme
	themeIcons[theme.Id] = make(map[string][]*Icon)
}

type Icon struct {
	resource.AbstractResource
	Name        string
	Theme       string
	Context     string
	Type        string
	MinSize     uint32
	MaxSize     uint32
	Path        string
	themeSubdir string
}

type Theme struct {
	resource.AbstractResource
	Id       string
	Name     string
	Comment  string
	Inherits []string
	Dirs     map[string]IconDir
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

type PngSvgPair struct {
	Png *Icon
	Svg *Icon
}
