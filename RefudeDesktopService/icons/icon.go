// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"sort"
	"strings"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var themes map[string]*Theme
var themeLock sync.Mutex

var themeIcons map[string]map[string][]*Icon
var otherIcons map[string]*Icon
var iconsByPath map[string]*Icon
var iconLock sync.Mutex

func GetThemes() []resource.Resource {
	themeLock.Lock()
	defer themeLock.Unlock()

	var themes = make([]resource.Resource, 0, len(themes))
	for _, theme := range themes {
		themes = append(themes, theme)
	}
	sort.Sort(resource.ResourceList(themes))
	return themes
}

func getTheme(themeId string) *Theme {
	themeLock.Lock()
	defer themeLock.Unlock()
	return themes[themeId]
}

func GetTheme(path string) *Theme {
	if !strings.HasPrefix(path, "/icontheme/") {
		return nil
	} else {
		return getTheme(string(path[len("/icontheme/"):]))
	}
}

func GetIcons() []resource.Resource {
	iconLock.Lock()
	defer iconLock.Unlock()

	var icons = make([]resource.Resource, 0, 1000 /*?*/)
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
	sort.Sort(resource.ResourceList(icons))
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
	// We prefer png's or svg's over xpm's
	if icon.Type == "xpm" {
		if _, ok := otherIcons[icon.Name]; ok {
			return
		}
	}
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
	resource.GenericResource
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
	resource.GenericResource
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
