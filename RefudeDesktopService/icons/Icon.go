// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"crypto/sha1"
	"fmt"
	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/server"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
)

type Icon struct {
	resource.AbstractResource
	Name        string
	Theme       string
	Context     string
	Type        string
	MinSize     uint32
	MaxSize     uint32
	Path        string
	themeSubDir string
}

type Theme struct {
	resource.AbstractResource
	Id       string
	Name     string
	Comment  string
	Inherits []string
	iconDirs map[string]IconDir
	icons    map[string][]*Icon
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

// Inspired by pseudocode example in
// https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html#icon_lookup
func (t *Theme) FindIcon(name string, size uint32, iconType string) *Icon {
	shortestDistanceSoFar := uint32(math.MaxUint32)
	var candidate *Icon = nil

	for _, icon := range t.icons[name] {
		if icon.Type != iconType {
			continue
		}

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
	mutex       sync.Mutex
	themes      map[string]*Theme
	otherIcons  map[string]*Icon
	iconsByPath map[resource.StandardizedPath]*Icon

	// Icons not yet associated with a theme (We may find the icon before the theme index)
	strayIcons map[string]map[string][]*Icon
}

func MakeIconCollection() *IconCollection {
	var ic = &IconCollection{}
	ic.themes = make(map[string]*Theme)
	ic.otherIcons = make(map[string]*Icon)
	ic.iconsByPath = make(map[resource.StandardizedPath]*Icon)
	ic.strayIcons = make(map[string]map[string][]*Icon)
	return ic
}

func (ic *IconCollection) SetThemes(themes map[string]*Theme) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	ic.themes = themes
	for path, icon := range ic.iconsByPath {
		if icon.Theme != "" {
			delete(ic.iconsByPath, path)
		}
	}

	for _, theme := range ic.themes {
		for _, icons := range theme.icons {
			for _, icon := range icons {
				ic.iconsByPath[resource.Standardize("/icon"+icon.Path)] = icon
			}
		}
	}
}

func (ic *IconCollection) SetOtherIcons(otherIcons map[string]*Icon) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	ic.otherIcons = otherIcons

	for path, icon := range ic.iconsByPath {
		if icon.Theme == "" {
			delete(ic.iconsByPath, path)
		}
	}

	for _, icon := range ic.otherIcons {
		ic.iconsByPath[resource.Standardize("/icon"+icon.Path)] = icon
	}
}

func (ic *IconCollection) GetThemes() []interface{} {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	var themes = make([]interface{}, 0, len(ic.themes))
	for _, theme := range ic.themes {
		themes = append(themes, theme)
	}
	return themes
}

func (ic *IconCollection) GetIcons() []interface{} {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

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
}

func (ic *IconCollection) GetIcon(path resource.StandardizedPath) *Icon {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	return ic.iconsByPath[path]
}

func (ic *IconCollection) GetIconByName(r *http.Request) (*PngSvgPair, bool) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	if names, size, theme, ok := extractNameSizeAndTheme(r.URL.Query()); !ok {
		return nil, false
	} else {
		fmt.Println("Looking for", names, size, theme)
		fmt.Println("Themes are:")
		for id, theme := range ic.themes {
			fmt.Println(id, theme.Name, "-", len(theme.icons), "icons");
		}
		var pngIcon = ic.findIcon(theme, names, size, "png")
		var svgIcon = ic.findIcon(theme, names, size, "svg")
		if (pngIcon != nil || svgIcon != nil) {
			return &PngSvgPair{pngIcon, svgIcon}, true
		} else {
			return nil, true
		}
	}

}

func extractNameSizeAndTheme(query map[string][]string) ([]string, uint32, string, bool) {
	if len(query["name"]) < 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return make([]string, 0), 0, "", false
	}

	var names = query["name"]
	var iconSize = uint32(32)
	var theme = "hicolor"

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
	if r.URL.Path == "/iconthemes" {
		filterAndServe(w, r, ic.GetThemes())
	} else if strings.HasPrefix(r.URL.Path, "/icontheme/") {
		if theme, ok := ic.themes[r.URL.Path[len("/icontheme/"):]]; ok {
			respond(w, r, theme, "", "")
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else if r.URL.Path == "/icons" || r.URL.Path == "/icons/" {
		filterAndServe(w, r, ic.GetIcons())
	} else if strings.HasPrefix(r.URL.Path, "/icon/") {
		fmt.Println("Looking for", r.URL.RawPath)
		if icon := ic.GetIcon(resource.StandardizedPath(r.URL.RawPath)); icon != nil {
			fmt.Println("Found", icon)
			if icon.Type == "png" {
				respond(w, r, icon, icon.Path, "")
			} else {
				respond(w, r, icon, "", icon.Path)
			}
		} else {
			fmt.Println("Nix gefunden")
			w.WriteHeader(http.StatusNotFound)
		}

	} else if r.URL.Path == "/iconsearch" {
		fmt.Println("/iconsearc..")
		if pngSvgPair, ok := ic.GetIconByName(r); !ok {
			requests.ReportUnprocessableEntity(w, errors.New("Invalid query parameters"))
		} else if pngSvgPair == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			var pngPath, svgPath string
			if pngSvgPair.Png != nil {
				pngPath = pngSvgPair.Png.Path
			}
			if pngSvgPair.Svg != nil {
				svgPath = pngSvgPair.Svg.Path
			}
			respond(w, r, pngSvgPair, pngPath, svgPath)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (ic *IconCollection) findIcon(themeId string, names []string, size uint32, iconType string) *Icon {
	var visited = make(map[string]bool) // Lists visited theme ids
	var toVisit = make([]string, 1, 10);
	toVisit[0] = themeId;
	for i := 0; i < len(toVisit); i++ {
		var themeId = toVisit[i]
		if theme, ok := ic.themes[themeId]; ok {
			for _, name := range names {
				if icon := theme.FindIcon(name, size, iconType); icon != nil {
					return icon
				}
			}
			visited[themeId] = true
			for _, parentId := range theme.Inherits {
				if ! visited[parentId] {
					toVisit = append(toVisit, parentId)
				}
			}
		}
	}

	var hicolorTheme = ic.themes["hicolor"];
	for _, name := range names {
		if icon := hicolorTheme.FindIcon(name, size, iconType); icon != nil {
			return icon
		}
	}

	for _, name := range names {
		if icon, ok := ic.otherIcons[name]; ok && icon.Type == iconType {
			return icon
		}

	}

	return nil
}

func (ic *IconCollection) addIcon(icon *Icon) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()

	if icon.Theme != "" {
		if theme, ok := ic.themes[icon.Theme]; ok {
			if iconDir, ok := theme.iconDirs[icon.themeSubDir]; ok {
				icon.MinSize, icon.MaxSize, icon.Context = iconDir.MinSize, iconDir.MaxSize, iconDir.Context
				theme.icons[icon.Name] = append(theme.icons[icon.Name], icon)
				ic.iconsByPath[icon.GetSelf()] = icon
			} else {
				log.Println("Unable to place icon", icon)
			}

		} else {
			strayIconsForTheme, ok := ic.strayIcons[icon.Theme]
			if !ok {
				strayIconsForTheme = make(map[string][]*Icon)
				ic.strayIcons[icon.Theme] = strayIconsForTheme
			}

			strayIconsForTheme[icon.Name] = append(strayIconsForTheme[icon.Name], icon)
		}
	} else {
		ic.otherIcons[icon.Name] = icon
		ic.iconsByPath[icon.GetSelf()] = icon
	}

}

func (ic *IconCollection) addTheme(theme *Theme) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	if _, ok := ic.themes[theme.Id]; ok {
		log.Print("Ignoring ", theme.Id, ", have it already.")
	}

	ic.themes[theme.Id] = theme

	if strayIconsForTheme, ok := ic.strayIcons[theme.Id]; ok {
		for name, icons := range strayIconsForTheme {
			for _, icon := range icons {
				if iconDir, ok := theme.iconDirs[icon.themeSubDir]; ok {
					icon.MinSize, icon.MaxSize, icon.Context = iconDir.MinSize, iconDir.MaxSize, iconDir.Context
					theme.icons[name] = append(theme.icons[name], icon)
				} else {
					log.Println("Unable to place icon", icon)
				}
			}
		}
		delete(ic.strayIcons, theme.Id)
	}
}

func filterAndServe(w http.ResponseWriter, r *http.Request, resources []interface{}) {
	if matcher, err := requests.GetMatcher2(r); err != nil {
		requests.ReportUnprocessableEntity(w, err)
	} else {
		var toServe []interface{}
		if matcher != nil {
			toServe = make([]interface{}, 0, len(resources))
			for _, res := range resources {
				if matcher(res) {
					toServe = append(toServe, res)
				}
			}
		} else {
			toServe = resources
		}

		var json = server.ToJSon(toServe)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", fmt.Sprintf("\"%x\"", sha1.Sum(json)))
		w.Write(json)
	}
}

func respond(w http.ResponseWriter, r *http.Request, res interface{}, pngPath string, svgPath string) {
	var offers = make([]string, 0, 3)

	offers = append(offers, "application/json")

	if pngPath != "" {
		offers = append(offers, "image/png")
	}

	if svgPath != "" {
		offers = append(offers, "image/svg")
	}

	if mimetype, err := requests.Negotiate(r, offers...); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else if mimetype == "application/json" {
		var json = server.ToJSon(res)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", fmt.Sprintf("\"%x\"", sha1.Sum(json)))
		w.Write(json)
	} else if mimetype == "image/png" {
		http.ServeFile(w, r, pngPath)
	} else if mimetype == "image/svg" {
		http.ServeFile(w, r, svgPath)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}
