// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"log"
	"path/filepath"
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type (
	themeIconImage struct {
		themeId  string
		iconDir  string
		iconName string
		path     string
	}

	sessionIconImage struct {
		name string
		size uint32
	}

	otherIconImage struct {
		name string
		path string
	}
)

var (
	lock              = &sync.Mutex{} // Global lock. TODO: lock with finer granularity
	themes            = make(map[string]*IconTheme)
	themeIcons        = make(map[string]map[string][]IconImage) // themeid -> iconname -> []IconImage
	wannabeThemeIcons = []themeIconImage{}                      // Icon images that seem to belong to a theme, but we seen the theme.index file yet.
	sessionIcons      = make(map[string][]IconImage)            // name -> []IconImage
	otherIcons        = make(map[string]IconImage)              // name -> IconImage
)

func AddBaseDir(baseDir string) {
	var foundThemes = false
	if indexFilePaths, err := filepath.Glob(baseDir + "/*/index.theme"); err != nil {
		log.Println(err)
	} else {
		for _, indexFilePath := range indexFilePaths {
			if theme, ok := readTheme(indexFilePath); ok {
				if _, ok = themes[theme.Id]; !ok {
					theme.Links = resource.MakeLinks("/icontheme/"+theme.Id, "icontheme")
					themes[theme.Id] = theme
					themeIcons[theme.Id] = make(map[string][]IconImage)
					foundThemes = true
				}
			}
		}
	}

	if foundThemes {
		mapThemeResources()
	}

	scanBaseDir(baseDir)
}

func mapThemeResources() {
	var themeResources = make(map[string]interface{})
	for themeId, theme := range themes {
		themeResources["/icontheme/"+themeId] = &(*theme)
	}
	themeResources["/iconthemes"] = resource.ExtractResourceList(themeResources)
	resource.MapCollection(&themeResources, "iconthemes")

	resource.MapSingle("/icon", &IconResource{})
}

func addThemeIconImage(themeId string, iconDir string, iconName string, path string) {
	lock.Lock()
	defer lock.Unlock()

	if theme, ok := themes[themeId]; !ok {
		wannabeThemeIcons = append(wannabeThemeIcons, themeIconImage{themeId, iconDir, iconName, path})
	} else if iconDir, ok := theme.Dirs[iconDir]; !ok {
		// Ignore
	} else {
		themeIcons[themeId][iconName] = append(themeIcons[themeId][iconName], IconImage{
			Context: iconDir.Context,
			MinSize: iconDir.MinSize,
			MaxSize: iconDir.MaxSize,
			Path:    path,
		})
	}
}

func AddOtherIcon(name string, path string) {
	lock.Lock()
	defer lock.Unlock()

	otherIcons[name] = IconImage{Path: path}
}
