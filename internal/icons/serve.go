// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package icons

import (
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/image"
	"github.com/surlykke/refude/internal/lib/log"
	"github.com/surlykke/refude/internal/lib/repo"
	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/lib/xdg"
)

var ThemeMap = repo.MakeSynkMap[string, *IconTheme]()

func Run() {
	collectThemes()
	collectIcons()
}

func GetHandler(name string, size uint32) response.Response {
	var iconFilePath = FindIcon(name, size)
	if iconFilePath == "" {
		return response.NotFound()
	}

	var (
		contentType string = "image/png"
		bytes       []byte
		err         error
	)
	if strings.HasSuffix(iconFilePath, ".xpm") {
		if bytes = getPngFromXpm(iconFilePath); bytes == nil {
			return response.NotFound()
		}
	} else if bytes, err = os.ReadFile(iconFilePath); err != nil {
		return response.NotFound()
	}

	if strings.HasSuffix(iconFilePath, ".svg") {
		contentType = "image/svg+xml"
	}
	return response.Image(contentType, bytes)
}

func getPngFromXpm(filePath string) []byte {
	xpmCacheLock.Lock()
	defer xpmCacheLock.Unlock()
	var pngBytes []byte
	var ok bool
	if pngBytes, ok = xpmCache[filePath]; !ok {
		var err error
		if pngBytes, err = image.Xpmfile2png(filePath); err != nil {
			log.Warn("Error converting xpm file:", err)
			xpmCache[filePath] = nil
		}
	}
	return pngBytes
}

var xpmCache = make(map[string][]byte)
var xpmCacheLock sync.Mutex

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
