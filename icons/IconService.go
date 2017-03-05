package main

import (
	"net/http"
	"sync"
	"net/url"
	"fmt"
)

type IconService struct {
	mutex  sync.RWMutex
	themes Themes
}

func (is IconService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		is.GET(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (is IconService) GET(w http.ResponseWriter, r* http.Request) {
	qParms := r.URL.Query()

	if len(qParms["name"]) != 1 || len(qParms["themeName"]) > 1 || len(qParms["size"]) > 1 {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	iconName := qParms["name"][0]
	themeName := getFirstOr(qParms, "theme", is.defaultTheme())

	iconSize,ok := readUint32(getFirstOr(qParms, "size", "32"))
	if !ok {
		iconSize = 32
	}

	is.mutex.RLock()
	themesCopy := is.themes
	is.mutex.RUnlock()

	if icon, ok := themesCopy.FindIcon(themeName, iconSize, iconName); ok {
		fmt.Println("Serving: ", icon.Path)
		http.ServeFile(w, r, icon.Path)
	}  else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func getFirstOr(values url.Values, key string, fallBack string) string {
	if valsForKey, ok := values[key]; ok {
		if len(valsForKey) > 0 {
			return valsForKey[0]
		}
	}
	return fallBack
}

func (is IconService) defaultTheme() string {
    return "oxygen" // FIXME
}

func (is *IconService) start() {
	is.update()
	http.ListenAndServe(":8000", is)
}

func (is *IconService) update() {
	is.mutex.Lock()
	defer is.mutex.Unlock()
	is.themes = ReadThemes()
}
