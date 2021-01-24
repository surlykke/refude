// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icons

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/surlykke/RefudeServices/lib/image"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

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

func xpm2Png(xpmFile string) ([]byte, error) {
	if xpmBytes, err := ioutil.ReadFile(xpmFile); err != nil {
		return nil, err
	} else {
		return image.Xpm2png(xpmBytes)
	}
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
