// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package icons

import (
	"errors"
	"fmt"
	"os"
	gopath "path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/surlykke/refude/internal/lib/image"
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/xdg"
)

func collectThemeIcons() {
	var searchOrder = determineSearchOrder()
	for _, themeId := range searchOrder {
		collectIconsFromTheme(themeId)
	}
}

func collectIconsFromTheme(themeId string) {
	var theme, _ = resource.Get[*IconTheme](themePathPrefix + themeId)
	for _, basedir := range xdg.IconBasedirs {
		for _, themeDir := range theme.Dirs {
			var glob = basedir + "/" + themeId + "/" + themeDir.Path + "/*"
			if filePathsInThemeDir, err := filepath.Glob(glob); err == nil {
				for _, filePath := range filePathsInThemeDir {
					if !looksLikeAnIconfile(filePath) {
						continue
					} else {
						addIconPath(iconName(filePath), themeId, filePath, themeDir.MinSize, themeDir.MaxSize)
					}
				}
			}
		}
	}
}

func collectOtherIcons() {
	var dirsToLookAt = make([]string, 0, len(xdg.IconBasedirs)+1)
	dirsToLookAt = append(dirsToLookAt, xdg.IconBasedirs...)
	dirsToLookAt = append(dirsToLookAt, xdg.PixmapDir)

	for _, dir := range dirsToLookAt {
		if filePathsInDir, err := filepath.Glob(dir + "/*"); err == nil {
			for _, filePath := range filePathsInDir {
				if !looksLikeAnIconfile(filePath) {
					continue
				} else {
					addIconPath(iconName(filePath), "", filePath, 1, 1)
				}
			}
		}
	}
}

func looksLikeAnIconfile(filePath string) bool {
	var ext = gopath.Ext(filePath)
	return ext == ".png" || ext == ".svg" || ext == ".xpm"
}

func iconName(filePath string) string {
	return gopath.Base(filePath[:len(filePath)-4])
}

func addIconPath(name, theme, osPath string, minSize, maxSize uint32) error {
	if tmp, err := processPath(osPath); err != nil {
		return err
	} else {
		osPath = tmp
	}
	var icon Icon
	if tmp, ok := resource.Get[*Icon](iconPrefix + name); !ok {
		icon =  Icon{Base: resource.MakeBase(iconPrefix+name, name, "", ""), Name: name, Theme: theme}
	} else if tmp.Theme == theme {
		icon = *tmp
	} else {
		return nil
	}

	if !icon.HasPath(osPath) {
		icon.Imgs = append(icon.Imgs, Img{OsPath: osPath, MinSize: minSize, MaxSize: maxSize})
	}
	resource.Put(&icon)
	return nil
}

func processPath(osPath string) (string, error) {
	if realPath, err := filepath.EvalSymlinks(osPath); err != nil {
		return "", err
	} else if strings.HasSuffix(realPath, ".xpm") {
		return convertToPng(realPath)
	} else {
		return realPath, nil
	}
}

func convertToPng(filePath string) (string, error) {
	if bytes, err := image.Xpmfile2png(filePath); err != nil {
		return "", err
	} else {
		var newPath = fmt.Sprintf("%s/%s.png", xdg.SessionFilesDir, image.HashName(bytes))
		if _, err := os.Stat(newPath); err == nil {
			return newPath, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return "", err
		} else if err := os.WriteFile(newPath, bytes, 0700); err != nil {
			os.Remove(newPath)
			return "", err
		} else {
			return newPath, nil
		}
	}
}

func determineSearchOrder() []string {
	var searchOrder = make([]string, 0, 4)
	var walker func(themeId string)
	walker = func(themeId string) {
		if themeId != "" && themeId != "hicolor" && !slices.Contains(searchOrder, themeId) {
			if theme, ok := resource.Get[*IconTheme](themePathPrefix + themeId); ok {
				searchOrder = append(searchOrder, themeId)
				for _, parent := range theme.Inherits {
					walker(parent)
				}
			}
		}
	}
	walker(determineDefaultThemeId())
	searchOrder = append(searchOrder, "hicolor") // hicolor is the general fall back
	fmt.Println("Icon theme search order:", searchOrder)
	return searchOrder
}

func determineDefaultThemeId() string {
	var iconThemeDefPattern = regexp.MustCompile(`gtk-icon-theme-name=(\S+)`)

	if defaultThemeId, ok := os.LookupEnv("REFUDE_ICON_THEME"); ok {
		return defaultThemeId
	} else {
		for _, iniFile := range []string{
			xdg.ConfigHome + "/gtk-4.0/settings.ini",
			"/etc/gtk-4.0/settings.ini",
			xdg.ConfigHome + "/gtk-3.0/settings.ini",
			"/etc/gtk-3.0/settings.ini",
			xdg.ConfigHome + "/gtk-2.0/settings.ini",
			"/etc/gtk-2.0/settings.ini",
			xdg.Home + "/.gtkrc-2.0", "/etc/gtk-2.0/gtkrc"} {

			if bytes, err := os.ReadFile(iniFile); err == nil {
				if matches := iconThemeDefPattern.FindStringSubmatch(string(bytes)); matches != nil {
					return matches[1]
				}
			}
		}
	}

	return ""
}
