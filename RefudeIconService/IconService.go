/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"net/http"
	"sync"
	"fmt"
)

type IconService struct {
	mutex  sync.RWMutex
	themes Themes
}

func (is IconService) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	switch r.URL.Path {
	case "/ping":
		if r.Method == "GET" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	case "/icon":
		fmt.Println("Serving ", r.URL.Path)
		if r.Method == "GET" {
			is.mutex.RLock()
			themesCopy := is.themes
			is.mutex.RUnlock()
			if names, size, theme, ok := is.extractNameSizeAndTheme(r.URL.Query()); !ok {
				w.WriteHeader(http.StatusUnprocessableEntity)
			} else {
				for _, name := range names {
					if icon, ok := themesCopy.FindIcon(theme, size, name); ok {
						fmt.Println("Serving: ", icon.Path)
						http.ServeFile(w, r, icon.Path)
						return
					}
				}
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}


func (is IconService) extractNameSizeAndTheme(query map[string][]string) ([]string, uint32, string, bool) {
	if len(query["name"]) < 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return make([]string, 0), 0, "", false
	}

	name := query["name"]
	iconSize := uint32(32)
	theme := is.defaultTheme()

	if len(query["size"]) > 0 {
		var ok bool
		if iconSize, ok = readUint32(query["size"][0]); !ok {
			return make([]string, 0), 0, "", false
		}
	}

	if len(query["theme"]) > 0 {
		theme = query["theme"][0]
	}

	return name, iconSize, theme, true
}


func (is IconService) defaultTheme() string {
    return "oxygen" // TODO
}



func (is *IconService) update() {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	is.themes = ReadThemes()
}
