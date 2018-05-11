// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package main

import (
	"fmt"
	"net/http"
	"sync"
	"path/filepath"
	"log"
	"os"
	"github.com/surlykke/RefudeServices/lib/xdg"
	"github.com/fsnotify/fsnotify"
)

var mutex  sync.RWMutex
var themes map[string]Theme
var fallbackIcons map[string]Icon


func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if "/icon" == r.URL.Path {
		if r.Method == "GET" {
			names, size, theme, ok := extractNameSizeAndTheme(r.URL.Query());
			fmt.Println("names, size, theme: ", names, size, theme)
			if !ok {
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			icon, ok := findIcon(names, size, theme)
			if ok {
				fmt.Println("Found icon: ", icon)
				http.ServeFile(w, r, icon.FindImage(size).Path)
			} else {
				fmt.Println("Icon not found")
				w.WriteHeader(http.StatusNotFound)
			}
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

func findIcon(names []string, size uint32, startThemeName string) (Icon, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	if startTheme, ok := themes[startThemeName]; !ok {
		fmt.Println("No startTheme")
		return Icon{}, false
	} else {
		for _,themeName := range startTheme.SearchOrder {
			fmt.Println("Searching in: ", themeName)
			theme := themes[themeName]
			for _,name := range names {
				if icon, ok := theme.Icons[name]; ok {
					return icon, true
				}
			}
		}

		for _,name := range names {
			fmt.Println("Searching fallbackIcons")
			if icon, ok := fallbackIcons[name]; ok {
				return icon, ok
			}
		}
		return Icon{}, false
	}
}


func extractNameSizeAndTheme(query map[string][]string) ([]string, uint32, string, bool) {
	if len(query["name"]) < 1 || len(query["themeName"]) > 1 || len(query["size"]) > 1 {
		return make([]string, 0), 0, "", false
	}

	name := query["name"]
	iconSize := uint32(32)
	theme := defaultTheme()

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

func defaultTheme() string {
	return "oxygen" // FIXME
}


func run() {
	fmt.Println("IconService, run")
	var session_icons_dir = xdg.RuntimeDir + "/org.refude.icon-service-session-icons"
	var marker_path = session_icons_dir + "/marker"

	if err := os.MkdirAll(session_icons_dir, os.ModePerm); err != nil {
		panic(err)
	} else if file,err := os.Create(marker_path); err != nil {
		panic(err)
	} else {
		file.Close()
	}

	if watcher, err := fsnotify.NewWatcher(); err != nil {
		panic(err)
	} else {
		defer watcher.Close()
		if err = watcher.Add(marker_path); err != nil {
			panic(err)
		}

		var tmpThemes = collectThemes()
		var tmpFallbackIcons = make(map[string]Icon)

		for _,searchDir := range getSearchDirectories() {
			collect(tmpThemes, tmpFallbackIcons, searchDir)
		}
		collect(tmpThemes, tmpFallbackIcons, session_icons_dir)
		
		mutex.Lock()
		themes = tmpThemes
		fallbackIcons = tmpFallbackIcons
		mutex.Unlock()

		for {
			select {
			case <-watcher.Events:
				recollect(session_icons_dir)
			case err := <-watcher.Errors:
				log.Println(err)
			}
		}
	}

}

func collectThemes() map[string]Theme {
	themes := make(map[string]Theme)

	for _, searcDir := range searchDirectories {
		indexThemeFilePaths, err := filepath.Glob(searcDir + "/" + "*" + "/index.theme")
		if err != nil {
			panic(err)
		}

		for _, indexThemeFilePath := range indexThemeFilePaths {
			fmt.Println("Looking at: ", indexThemeFilePath)
			themeId := filepath.Base(filepath.Dir(indexThemeFilePath))
			if _, ok := themes[themeId]; !ok {
				if theme, err := readIndexTheme(themeId, indexThemeFilePath); err == nil {
					themes[themeId] = theme
				} else {
					log.Println("Error reading index.theme: ", err)
				}

			}
		}
	}

	for themeId, theme := range themes {
		searchOrder := getAncestors(themeId, make([]string, 0), themes)
		searchOrder = append(searchOrder, "hicolor")
		theme.SearchOrder = searchOrder
		themes[themeId] = theme
	}


	return themes
}

func collect(themes map[string]Theme, fallbackIcons map[string]Icon, searchDir string) {
	for _, theme := range themes {
		for _, iconDir := range theme.IconDirs {
			iconDirPath := searchDir + "/" + theme.Id + "/" + iconDir.Path
			if _, err := os.Stat(iconDirPath); err != nil {
				continue
			}
			collectIcons(theme.Icons, iconDirPath, iconDir)
		}
	}

	dummyIconDir := IconDir{}
	collectIcons(fallbackIcons, searchDir, dummyIconDir)

}



func recollect(searchDir string) {
	fmt.Println("Recollecting from", searchDir)
	mutex.Lock()
	defer mutex.Unlock()

	collect(themes, fallbackIcons, searchDir)
}

func collectIcons(icons map[string]Icon, iconDirPath string, iconDir IconDir) {
	for _, ending := range []string{"png", "svg", "xpm"} {
		imagePaths, err := filepath.Glob(iconDirPath + "/*." + ending)
		if err != nil {
			panic(err)
		}

		for _, imagePath := range imagePaths {
		    var image = Image{iconDir.Context, iconDir.MinSize, iconDir.MaxSize, imagePath}
			var iconName = filepath.Base(imagePath[0 : len(imagePath)-4])
			if iconName == "chrome_app_indicator_1_1" {
				fmt.Println("imagePath: ", imagePath)
			}
			var icon, ok = icons[iconName]
			if !ok {
				icon = Icon{iconName, []Image{image}}
			} else {
				icon.AddImage(image, false)
			}
			icons[iconName] = icon
		}
	}
}

// With xdg dirs at their default values, we search directories in this order:
// $HOME/.icon, $HOME/.local/share/icons, /usr/local/share/icons, /usr/share/icons, /usr/share/pixmap
// Ie. 'more local' takes precedence. eg:
// If both $HOME/.local/share/icons/hicolor/22x22/apps/myIcon.png and /usr/share/icons/hicolor/22x22/apps/myIcon.png
// exists, we prefer the one under $HOME/.local
func getSearchDirectories() []string {
	searchDirs := []string{xdg.Home + "/.icons", xdg.DataHome + "/icons"}
	for _, datadir := range reverse(xdg.DataDirs) {
		searchDirs = append(searchDirs, datadir+"/icons")
	}
	searchDirs = append(searchDirs, xdg.RuntimeDir + "/refude-icons")
	searchDirs = append(searchDirs, "/usr/share/pixmaps")
	return searchDirs
}

