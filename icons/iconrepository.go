// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package icons

import (
	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/slice"
)

type nameAndSize struct {
	name string
	size uint32
}

type addIconRequest struct {
	name string
	size uint32
	png  []byte
}

type iconFilePathRequest struct {
	name   string
	size   uint32
	result chan string
}

var themeRepo = repo.MakeRepo[*IconTheme]()
var addIconRequests = make(chan addIconRequest)
var addIconFileRequests = make(chan string)
var addBaseDirRequests = make(chan string)
var iconFilePathRequests = make(chan iconFilePathRequest)

func Run() {
	var addedFiles = make(map[string]bool) // We only serve icons by absolute path if found here
	var addedSessionIcons = make(map[string]bool)

	var itc, err = buildIconThemeCollection()
	if err != nil {
		panic(err)
	}

	themeRepo.RemoveAll()
	for _, theme := range itc.allThemes {
		themeRepo.Put(theme)
	}
	// -------------------

	var repoRequests = repo.MakeAndRegisterRequestChan()
	for {
		select {
		case req := <-repoRequests:
			themeRepo.DoRequest(req)
		case req := <-addIconRequests:
			if req.size > 0 {
				itc.writeSessionHicolorIcon(req.name, req.size, req.png)
			} else {
				itc.writeSessionOtherIcon(req.name, req.png)
			}
			addedSessionIcons[req.name] = true
		case req := <-addIconFileRequests:
			addedFiles[req] = true
		case req := <-iconFilePathRequests:
			req.result <- itc.locateIcon(req.name, req.size)
		case path := <-addBaseDirRequests:
			itc.basedirs = slice.AppendIfNotThere(itc.basedirs, path)
		}
	}
}

func FindIconPath(name string, size uint32) string {
	var ifpr = iconFilePathRequest{
		name:   name,
		size:   size,
		result: make(chan string),
	}
	iconFilePathRequests <- ifpr
	return <-ifpr.result
}

func AddARGBIcon(argbIcon image.ARGBIcon) string {
	var iconName = image.ARGBIconHashName(argbIcon)

	for _, pixMap := range argbIcon.Images {
		if pixMap.Width == pixMap.Height { // else ignore
			if png, err := pixMap.AsPng(); err != nil {
				log.Warn("Unable to convert image", err)
			} else {
				addIconRequests <- addIconRequest{name: iconName, size: pixMap.Height, png: png}
			}
		}

	}
	return iconName
}

func AddFileIcon(filePath string) {
	addIconFileRequests <- filePath
}

func AddRawImageIcon(imageData image.ImageData) string {
	iconName := image.ImageDataHashName(imageData)
	if png, err := imageData.AsPng(); err != nil {
		log.Warn("Error converting image", err)
		return ""
	} else {
		addIconRequests <- addIconRequest{name: iconName, size: 0, png: png}
	}
	return iconName
}

func AddBasedir(path string) {
	addBaseDirRequests <- path
}
