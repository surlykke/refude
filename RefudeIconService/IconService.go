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
	switch r.Method {
	case "GET":
		is.mutex.RLock()
		themesCopy := is.themes
		is.mutex.RUnlock()
		if name, size, theme, ok := is.extractNameSizeAndTheme(r.URL.Query()); !ok {
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			if icon, ok := themesCopy.FindIcon(theme, size, name); ok {
				fmt.Println("Serving: ", icon.Path)
				http.ServeFile(w, r, icon.Path)
			}  else {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}


func (is IconService) extractNameSizeAndTheme(query map[string][]string) (string, uint32, string, bool) {
	if len(query["name"]) != 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return "", 0, "", false
	}

	name := query["name"][0]
	iconSize := uint32(32)
	theme := is.defaultTheme()

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


func (is IconService) defaultTheme() string {
    return "oxygen" // TODO
}



func (is *IconService) update() {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	is.themes = ReadThemes()
}
