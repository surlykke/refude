// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

/** Urls
 * TODO
 */

/**
 * Example naming of directory variables in this code:
 *
 * /usr/share/icons/oxygen/base/32x32/actions/
 * |-datadir-|
 * |----icondir----|
 * |--------themedir------|
 * |----------------themesubdir--------------|
 */

var refudeConvertedIconsDir string
var refudeSessionIconsDir string

func init() {
	refudeConvertedIconsDir = xdg.RuntimeDir + "/org.refude.converted-icons"
	if err := os.MkdirAll(refudeConvertedIconsDir, 0700); err != nil {
		panic(err)
	}
	refudeSessionIconsDir = xdg.RuntimeDir + "/org.refude.session-icons"
	if err := os.MkdirAll(refudeSessionIconsDir, 0700); err != nil {
		panic(err)
	}
}

var (
	baseDirSink = make(chan string)
)

func receiveIconDirs() {

	for baseDirPath := range baseDirSink {
		fmt.Println("Recieve", baseDirPath)
		baseDirPath = filepath.Clean(baseDirPath)
		if baseDir, err := os.Open(baseDirPath); err != nil {
			fmt.Println("Could not open", baseDirPath, "-", err)
		} else if fileInfos, err := baseDir.Readdir(-1); err != nil {
			fmt.Println("Could not read", baseDirPath, "-", err)
		} else {
			for _, fileInfo := range fileInfos {
				if strings.HasPrefix(fileInfo.Name(), ".") {
					continue
				} else if fileInfo.IsDir() {
					handleDirectory(baseDirPath, fileInfo.Name())
				} else {
					handleFile(baseDirPath, "", "", fileInfo)
				}
			}
		}

	}

}

func handleDirectory(baseDirPath string, dirName string) {
	var themeDirPath = baseDirPath + "/" + dirName

	filepath.Walk(themeDirPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err == nil && !fileInfo.Mode().IsDir() {
			var iconDir = ""
			if len(path) > len(themeDirPath)+len(fileInfo.Name())+2 {
				iconDir = path[len(themeDirPath)+1 : len(path)-len(fileInfo.Name())-1]
			}
			handleFile(baseDirPath, dirName, iconDir, fileInfo)
		}

		return nil
	})
}

func handleFile(baseDirPath string, themeId string, iconDir string, fileInfo os.FileInfo) {
	if fileInfo.Mode()&(os.ModeDevice|os.ModeNamedPipe|os.ModeSocket|os.ModeCharDevice) != 0 {
		return
	} else if hasOneOfSuffixes(fileInfo.Name(), ".xpm", ".png", ".svg") {
		var iconName = fileInfo.Name()[:len(fileInfo.Name())-4]
		var iconType = fileInfo.Name()[len(fileInfo.Name())-3:]
		var imageFilePath = filepath.Clean(baseDirPath + "/" + themeId + "/" + iconDir + "/" + fileInfo.Name())
		if iconType == "xpm" {
			var pngFilePath = filepath.Clean(refudeConvertedIconsDir + "/" + themeId + "/" + iconDir + "/" + iconName + ".png")
			if err := convertAndSave(imageFilePath, pngFilePath); err != nil {
				log.Println("Problem converting", imageFilePath, err)
				return
			}
			iconType = "png"
			imageFilePath = pngFilePath
		}

		if themeId != "" {
			themeIconImages <- themeIconImage{
				iconName: iconName,
				themeId:  themeId,
				iconDir:  iconDir,
				path:     imageFilePath,
			}
		} else {
			otherIconImages <- otherIconImage{
				name: iconName,
				path: imageFilePath,
			}
		}
	}
}

func hasOneOfSuffixes(s string, suffixes ...string) bool {
	for _, suffix := range suffixes {
		if strings.HasSuffix(s, suffix) {
			return true
		}
	}

	return false
}

func convertAndSave(pathToXpm string, pathToPng string) error {
	if xpmBytes, err := ioutil.ReadFile(pathToXpm); err != nil {
		return err
	} else {
		var pathToPngDir = filepath.Dir(pathToPng)
		if err := os.MkdirAll(pathToPngDir, 0700); err != nil {
			return err
		}
		if _, err := os.Stat(pathToPng); os.IsNotExist(err) {
			if pngBytes, err := image.Xpm2png(xpmBytes); err != nil {
				return err
			} else if err = ioutil.WriteFile(pathToPng, pngBytes, 0700); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}

		return nil
	}
}

type convertibleToPng interface {
	AsPng() ([]byte, error)
}

func saveAsPng(dir string, name string, image convertibleToPng) {
	if png, err := image.AsPng(); err != nil {
		log.Println("Error converting image to png:", err)
	} else {
		if err = os.MkdirAll(dir, os.ModePerm); err != nil {
			log.Println("Unable to create", dir, err)
		} else if err = ioutil.WriteFile(dir+"/"+name+".png", png, 0700); err != nil {
			log.Println("Unable to write file", err)
		}
	}

}
