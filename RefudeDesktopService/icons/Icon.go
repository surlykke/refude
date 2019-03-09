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
	"math"
	"net/http"
	"strings"
	"sync"
)

type img struct {
	MinSize uint32
	MaxSize uint32
	Path    string
}

type Icon struct {
	resource.AbstractResource
	Name    string
	Theme   string
	Context string
	Type    string
	img
}

type Theme struct {
	Id          string
	Name        string
	Comment     string
	Inherits    []string
	SearchOrder []string
	iconDirs    []IconDir
	icons       map[string][]*Icon
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
	mutex               sync.Mutex
	themes              map[string]*Theme
	otherIcons          map[string]*Icon
	iconsByPath         map[resource.StandardizedPath]*Icon
	server.PostNotAllowed
	server.PatchNotAllowed
	server.DeleteNotAllowed
}

func MakeIconCollection() *IconCollection {
	var ic = &IconCollection{}
	ic.themes = make(map[string]*Theme)
	ic.otherIcons = make(map[string]*Icon)
	ic.iconsByPath = make(map[resource.StandardizedPath]*Icon)
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
	fmt.Println("GetSingle, path:", path)
	return ic.iconsByPath[path]
}

func (ic *IconCollection) GetIconByName(r *http.Request) (*PngSvgPair, bool) {
	ic.mutex.Lock()
	defer ic.mutex.Unlock()
	if names, size, theme, ok := extractNameSizeAndTheme(r.URL.Query()); !ok {
		return nil, false
	} else {
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
	fmt.Println("GET", r.URL.Path)
	if r.URL.Path == "/iconthemes" {
		filterAndServe(w, r, ic.GetThemes())
	} else if r.URL.Path == "/icons" || r.URL.Path == "/icons/" {
		filterAndServe(w, r, ic.GetIcons())
	} else if strings.HasPrefix(r.URL.Path, "/icon/") {
		if icon := ic.GetIcon(resource.StandardizedPath(r.URL.Path)); icon != nil {
			if icon.Type == "png" {
				respond(w, r, icon, icon.Path, "")
			} else {
				respond(w, r, icon, "", icon.Path)
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}

	} else if r.URL.Path == "/iconsearch" {
		fmt.Println("Search pngSvgPair")
		if pngSvgPair, ok := ic.GetIconByName(r); !ok {
			requests.ReportUnprocessableEntity(w, errors.New("Invalid query parameters"))
		} else if pngSvgPair == nil {
			w.WriteHeader(http.StatusNotFound)
		} else {
			respond(w, r, pngSvgPair, pngSvgPair.Png.Path, pngSvgPair.Svg.Path)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func (IconCollection) HandledPrefixes() []string {
	return []string{"/icon"}
}

func (ic *IconCollection) findIcon(themeName string, names []string, size uint32, iconType string) *Icon {
	if _, ok := ic.themes[themeName]; ok {
		for _, themeName := range ic.themes[themeName].SearchOrder {
			if theme, ok := ic.themes[themeName]; ok {
				for _, name := range names {
					if icon := theme.FindIcon(name, size, iconType); icon != nil {
						return icon
					}
				}
			}
		}
	}

	for _, name := range names {
		if icon, ok := ic.otherIcons[name]; ok && icon.Type == iconType {
			return icon
		}

	}

	return nil
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

	if mimetype, err := negotiate(r, offers...); err != nil {
		w.WriteHeader(http.StatusBadRequest)
	} else if mimetype == "application/json" {
		fmt.Println("Serving application/json")
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
