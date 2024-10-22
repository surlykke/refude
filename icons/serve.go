// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package icons

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/icon"
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

func Run() {
	collectThemes()
	collectIcons()

	// TODO Recollect on changes
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/icon" {
		if r.Method == "GET" {
			if name := requests.GetSingleQueryParameter(r, "name", ""); name == "" {
				respond.UnprocessableEntity(w, fmt.Errorf("Query parameter 'name' must be given, and not empty"))
			} else if size, err := extractSize(r); err != nil {
				respond.UnprocessableEntity(w, err)
			} else if iconFilePath := FindIcon(name, size); iconFilePath == "" {
				respond.NotFound(w)
			} else {
				http.ServeFile(w, r, iconFilePath)
			}
		} else {
			respond.NotAllowed(w)
		}

	} else {
		respond.NotFound(w)
	}
}

func FindIcon(iconName string, size uint32) string {
	var icon = icon.Name(iconName)
	if iconPaths, ok := getIconPaths(icon); ok {
		return bestSizeMatch(iconPaths, size)
	} else if lastDash := strings.LastIndex(iconName, "-"); lastDash > -1 {
		/*
		   By the icon naming specification, dash ('-') seperates 'levels of specificity'. So given an icon name
		   'input-mouse-usb', the levels of specificy, and the names and order we search will be: 'input-mouse-usb',
		   'input-mouse' and 'input'. Here we prefer specificy over theme, ie. if 'input-mouse-usb' is found in an inherited theme, that
		   is preferred over 'input-mouse' in the default theme
		*/
		return FindIcon(iconName[0:lastDash], size)
	} else {
		return ""
	}

}

func AddARGBIcon(argbIcon image.ARGBIcon) icon.Name {
	var iconName = icon.Name(image.ARGBIconHashName(argbIcon))
	var iconPaths = make([]IconPath, 0, len(argbIcon.Images))
	for _, pixMap := range argbIcon.Images {
		if pixMap.Width == pixMap.Height { // else ignore
			if png, err := pixMap.AsPng(); err != nil {
				log.Warn("Unable to convert image", err)
			} else {
				var (
					size = pixMap.Width
					path = fmt.Sprintf("%s/%s_%d.png", sessionIconsDir, iconName, size)
				)
				if err := os.WriteFile(path, png, 0700); err != nil {
					log.Warn("Could not write", path, err)
				} else {
					iconPaths = append(iconPaths, IconPath{Path: path, MinSize: size, MaxSize: size})
				}
			}
		}
	}
	if len(iconPaths) > 0 {
		addSessionIcon(iconName, iconPaths)
		return iconName
	} else {
		return ""
	}
}

func AddFileIcon(filePath string) {
	addSessionIconSinglePath(icon.Name(filePath), filePath)
}

func AddRawImageIcon(imageData image.ImageData) icon.Name {
	iconName := icon.Name(image.ImageDataHashName(imageData))
	if png, err := imageData.AsPng(); err != nil {
		log.Warn("Error converting image", err)
		return ""
	} else {
		var path = fmt.Sprintf("%s/%s.png", sessionIconsDir, iconName)
		if err := os.WriteFile(path, png, 0700); err != nil {
			log.Warn("Could not write", path, err)
			return ""
		} else {
			addSessionIconSinglePath(iconName, path)
		}
	}
	return iconName
}

func AddPngIcon(png []byte) icon.Name {
	var iconName = icon.Name(image.HashName(png))
	var path = fmt.Sprintf("%s/%s.png", sessionIconsDir, iconName)
	if err := os.WriteFile(path, png, 0700); err != nil {
		log.Warn("Could not write", path, err)
		return ""
	} else {
		addSessionIconSinglePath(iconName, path)
		return iconName
	}
}

func AddBasedir(path string) {
	// FIXME
}

func UrlFromName(name string) string {
	if strings.Index(name, "/") > -1 {
		// So its a path..
		if strings.HasPrefix(name, "file:///") {
			name = name[7:]
		} else if strings.HasPrefix(name, "file://") {
			name = xdg.Home + "/" + name[7:]
		} else if !strings.HasPrefix(name, "/") {
			name = xdg.Home + "/" + name
		}

		AddFileIcon(name)
		// Maybe: Check that path points to iconfile..
	}
	if name != "" {
		return "http://localhost:7938/icon?name=" + url.QueryEscape(name)
	} else {
		return ""
	}
}

func bestSizeMatch(iconPaths []IconPath, size uint32) string {
	var shortestDistanceSoFar = uint32(math.MaxUint32)
	var path = ""
	for _, iconPath := range iconPaths {
		var distance uint32

		if iconPath.MinSize > size {
			distance = iconPath.MinSize - size
		} else if iconPath.MaxSize < size {
			distance = size - iconPath.MaxSize
		} else {
			distance = 0
		}

		if distance < shortestDistanceSoFar {
			shortestDistanceSoFar = distance
			path = iconPath.Path
			shortestDistanceSoFar = distance
		}
		if shortestDistanceSoFar == 0 {
			break
		}
	}
	return path
}

func extractSize(r *http.Request) (uint32, error) {
	var size uint32 = 32

	if len(r.URL.Query()["size"]) > 0 {
		if size64, err := strconv.ParseUint(r.URL.Query()["size"][0], 10, 32); err != nil {
			return 0, errors.New("Invalid size given:" + r.URL.Query()["size"][0])
		} else {
			size = uint32(size64)
		}
	}

	return size, nil
}
