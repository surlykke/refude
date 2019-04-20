// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"errors"
	"math"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/resource"
)

// Caller ensures r.URL.Path starts with '/icon/'
func ServeIcon(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if icon := getIcon(resource.StandardizedPath(r.URL.RawPath)); icon != nil {
		if icon.Type == "png" {
			respond(w, r, icon, icon.Path, "")
		} else {
			respond(w, r, icon, "", icon.Path)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}

}

// Caller ensures r.URL.Path == "/icon"
func ServeNamedIcon(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else if pngSvgPair, ok := getIconByName(r); !ok {
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

}

func getIcon(path resource.StandardizedPath) *Icon {
	iconLock.Lock()
	defer iconLock.Unlock()
	return iconsByPath[path]
}

func getIconByName(r *http.Request) (*PngSvgPair, bool) {
	if name, size, theme, ok := extractNameSizeAndTheme(r.URL.Query()); !ok {
		return nil, false
	} else {
		var pngIcon = findIcon(theme, name, size, "png")
		var svgIcon = findIcon(theme, name, size, "svg")
		if pngIcon != nil || svgIcon != nil {
			return &PngSvgPair{pngIcon, svgIcon}, true
		} else {
			return nil, true
		}
	}
}

func extractNameSizeAndTheme(query map[string][]string) (string, uint32, string, bool) {
	if len(query["name"]) < 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return "", 0, "", false
	}

	if len(query["name"]) != 1 {
		return "", 0, "", false

	}
	var name = query["name"][0]
	var iconSize = uint32(32)
	var theme = "oxygen"

	if len(query["size"]) > 0 {
		var ok bool
		if iconSize, ok = readUint32(query["size"][0]); !ok {
			return "", 0, "", false
		}
	}

	if len(query["theme"]) > 0 {
		theme = query["theme"][0]
	}

	return name, iconSize, theme, true
}

func findIcon(themeId string, name string, size uint32, iconType string) *Icon {
	var visited = make(map[string]bool)
	var toVisit = make([]string, 1, 10)
	toVisit[0] = themeId
	for i := 0; i < len(toVisit); i++ {
		var themeId = toVisit[i]
		if theme := getTheme(themeId); theme != nil {
			if icon := findIconInTheme(themeId, name, size, iconType); icon != nil {
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
		if icon := findIconInTheme("hicolor", name, size, iconType); icon != nil {
			return icon
		}
	}

	if icon, ok := otherIcons[name]; ok && icon.Type == iconType {
		return icon
	}

	return nil
}

// Inspired by pseudocode example in
// https://specifications.freedesktop.org/icon-theme-spec/icon-theme-spec-latest.html#icon_lookup
func findIconInTheme(themeId string, name string, size uint32, iconType string) *Icon {
	shortestDistanceSoFar := uint32(math.MaxUint32)
	var candidate *Icon = nil

	iconLock.Lock()
	defer iconLock.Unlock()
	// Caller ensures themeIcons[themeId] is not zero
	for _, icon := range themeIcons[themeId][name] {
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

func filterAndServe(w http.ResponseWriter, r *http.Request, resources []interface{}) {
	if matcher, err := requests.GetMatcher(r); err != nil {
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

		var json = resource.ToJSon(toServe)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", json.Etag)
		w.Write(json.Data)
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
		var json = resource.ToJSon(res)
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("ETag", json.Etag)
		w.Write(json.Data)
	} else if mimetype == "image/png" {
		http.ServeFile(w, r, pngPath)
	} else if mimetype == "image/svg" {
		http.ServeFile(w, r, svgPath)
	} else {
		w.WriteHeader(http.StatusNotAcceptable)
	}
}
