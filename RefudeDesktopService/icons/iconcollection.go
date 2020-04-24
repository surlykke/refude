// Copyright (c) 2017,2018,2019 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"
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
	if indexFilePaths, err := filepath.Glob(baseDir + "/*/index.theme"); err != nil {
		log.Println(err)
	} else {
		for _, indexFilePath := range indexFilePaths {
			if theme, ok := readTheme(indexFilePath); ok {
				addTheme(theme)
			}
		}
	}

	scanBaseDir(baseDir)
}

func addTheme(theme *IconTheme) {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := themes["/icontheme/"+theme.Id]; !ok {
		themes["/icontheme/"+theme.Id] = theme
		themeIcons[theme.Id] = make(map[string][]IconImage)
		var newWannabees = make([]themeIconImage, 0, len(wannabeThemeIcons))
		for _, wannabeIcon := range wannabeThemeIcons {
			if strings.Index(wannabeIcon.path, "bright") > -1 {
				fmt.Println("--------- consider", wannabeIcon.path)
			}
			if wannabeIcon.themeId == theme.Id {
				if iconDir, ok := theme.Dirs[wannabeIcon.iconDir]; ok {
					themeIcons[theme.Id][wannabeIcon.iconName] =
						append(themeIcons[theme.Id][wannabeIcon.iconName], IconImage{
							Context: iconDir.Context,
							MinSize: iconDir.MinSize,
							MaxSize: iconDir.MaxSize,
							Path:    wannabeIcon.path,
						})
				}
			} else {
				newWannabees = append(newWannabees, wannabeIcon)
			}
		}
		wannabeThemeIcons = newWannabees
	}
}

func addThemeIconImage(themeId string, iconDir string, iconName string, path string) {
	lock.Lock()
	defer lock.Unlock()

	if theme, ok := themes["/icontheme/"+themeId]; !ok {
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
