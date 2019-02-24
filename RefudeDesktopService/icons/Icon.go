// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"github.com/surlykke/RefudeServices/lib/server"
	"math"
	"net/http"
	"strings"
	"sync"
)

type Icon struct {
	Name    string
	Context string
	MinSize uint32
	MaxSize uint32
	Path    string
	Type    string
}

type Theme struct {
	Id          string
	Name        string
	Comment     string
	Inherits    []string
	searchOrder []string
	iconDirs    []IconDir
	icons       map[string][]*Icon
}

type IconDir struct {
	Path    string
	MinSize uint32
	MaxSize uint32
	Context string
}

// Inspired by pseudocode example in
// https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html#icon_lookup
func (t *Theme) FindIcon(name string, size uint32) *Icon {
	shortestDistanceSoFar := uint32(math.MaxUint32)
	var candidate *Icon = nil

	for _, icon := range t.icons[name] {
		var distance uint32
		if icon.MinSize > size {
			distance = icon.MinSize - size
		} else if icon.MaxSize < size {
			distance = size - icon.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			candidate = icon
		}
		if distance == 0 {
			break
		}
	}

	return candidate
}

type IconCollection struct {
	mutex sync.Mutex
	server.CachingJsonGetter
	themes     map[string]*Theme
	otherIcons map[string]*Icon
	server.PostNotAllowed
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func MakeIconCollection() *IconCollection {
	var ic = &IconCollection{}
	ic.CachingJsonGetter = server.MakeCachingJsonGetter(ic)
	ic.themes = make(map[string]*Theme)
	ic.otherIcons = make(map[string]*Icon)
	return ic
}

func (ic *IconCollection) SetThemes(themes map[string]*Theme) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	ic.CachingJsonGetter.Clear()
	ic.themes = themes
}

func (ic *IconCollection) SetOtherIcons(otherIcons map[string]*Icon) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	ic.CachingJsonGetter.Clear()
	ic.otherIcons = otherIcons
}

func (ic *IconCollection) GetCollection(r *http.Request) []interface{} {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	if r.URL.Path == "/icons" {
		var icons = make([]interface{}, 0, 1000 /*?*/)
		for _, theme := range ic.themes {
			for _, iconlist := range theme.icons {
				for _, icon := range iconlist {
					icons = append(icons, icon)
				}
			}
		}

		for _, icon := range ic.otherIcons {
			icons = append(icons, icon)
		}

		return icons
	} else if r.URL.Path == "/iconthemes" {
		var themes = make([]interface{}, 0, len(ic.themes))
		for _, theme := range ic.themes {
			themes = append(themes, theme)
		}

		return themes
	} else if strings.HasPrefix(r.URL.Path, "/icons/") {
		var themeName = r.URL.Path[len("/icons/"):]
		if theme, ok := ic.themes[themeName]; ok {
			var icons = make([]interface{}, 0, len(theme.icons))
			for _, iconlist := range theme.icons {
				for _, icon := range iconlist {
					icons = append(icons, icon)
				}
			}

			return icons
		} else {
			return nil
		}
	} else {
		return nil
	}
}

func (ic *IconCollection) GetSingle(r *http.Request) interface{} {
	return nil // FIXME
}

func extractNameSizeAndTheme(query map[string][]string) ([]string, uint32, string, bool) {
	if len(query["name"]) < 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return make([]string, 0), 0, "", false
	}

	var names = query["name"]
	var iconSize = uint32(32)
	var theme string

	if len(query["size"]) > 0 {
		var ok bool
		if iconSize, ok = readUint32(query["size"][0]); !ok {
			return make([]string, 0), 0, "", false
		}
	}

	if len(query["theme"]) > 0 {
		theme = query["theme"][0]
	}

	return names, iconSize, theme, true
}

func (ic *IconCollection) GET(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/iconsearch" {
		w.WriteHeader(http.StatusNotFound) // FIXME
	} else if strings.HasPrefix(r.URL.Path, "/icon/") {
		w.WriteHeader(http.StatusNotFound) // FIXME
	} else {
		ic.CachingJsonGetter.GET(w, r)
	}

}

func (IconCollection) HandledPrefixes() []string {
	return []string{"/icon"}
}

func (ic *IconCollection) findIconInTheme(themeName string, names []string, size uint32) *Icon {
	if _, ok := ic.themes[themeName]; ok {
		for _, themeName := range ic.themes[themeName].searchOrder {
			if theme, ok := ic.themes[themeName]; ok {
				for _, name := range names {
					if icon := theme.FindIcon(name, size); icon != nil {
						return icon
					}
				}
			}
		}
	}

	return nil
}
