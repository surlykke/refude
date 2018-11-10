// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package image

import (
	"github.com/surlykke/RefudeServices/lib/xdg"
	"image"
	"bytes"
	"image/color"
	"image/png"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"log"
	"crypto/sha1"
	"sync"
)

type PNGImg struct {
	Size    int
	pngData []byte
}

type PNGIcon struct {
	images []PNGImg
}



type Img struct {
	Width  int32
	Height int32
	/*
	 * Holds pixels in blocks of 4 bytes. Each block (from low address to high)
	 * the A,R,G and B component of the pixel
	 */
	Pixels []byte
}

func (img *Img) PixelAt(row int32, column int32) ([]byte, error) {
	if column < 0 || column > img.Width || row < 0 || row > img.Height {
		return nil, errors.New("Out of range")
	} else {
		pos := 4 * (row*img.Width + column)
		return img.Pixels[pos : pos+4], nil
	}
}

type Icon []Img



var hicolorMapSizes = map[int32]bool{
	16:  true,
	22:  true,
	24:  true,
	32:  true,
	36:  true,
	48:  true,
	64:  true,
	72:  true,
	96:  true,
	128: true,
	192: true,
	256: true,
	512: true,
}

var savedNames = make(map[string]bool)
var savedNamesLock sync.Mutex

// Returns false if name is registered already
func registerName(name string) bool {
	savedNamesLock.Lock()
	defer savedNamesLock.Unlock()
	if _, ok := savedNames[name]; ok {
		return false
	} else {
		savedNames[name] = true
		return true
	}
}

func SaveAsPngToSessionIconDir(argbIcon Icon) string {
	var sessionIconDir = xdg.RuntimeDir + "/org.refude.icon-service-session-icons/"
	var wroteSomething = false
	hash := sha1.New()
	for _, img := range argbIcon {
		hash.Write(img.Pixels)
	}
	var iconName = fmt.Sprintf("%X", hash.Sum(nil))
	if registerName(iconName) {

		for _, img := range argbIcon {
			if img.Height == img.Width && hicolorMapSizes[img.Height] {
				var destDir = fmt.Sprintf("%shicolor/%dx%d/apps", sessionIconDir, img.Width, img.Height)
				var destPath = destDir + "/" + iconName + ".png"
				if _, err := os.Stat(destPath); os.IsNotExist(err) {
					if err := os.MkdirAll(destDir, os.ModePerm); err != nil {
						continue
					}
					wroteSomething = wroteSomething || makeAndWritePng(img, destPath)
				} else {
					continue
				}
			} else {
				fmt.Println("Skip")
			}
		}

		if wroteSomething {
			if _, err := os.Create(sessionIconDir + "/marker"); err != nil {
				log.Println("Error updating marker:", err)
			}
		}
	}
	return iconName
}

func makeAndWritePng(img Img, path string) bool {
	pngData := image.NewRGBA(image.Rect(0, 0, int(img.Width), int(img.Height)))
	buf := bytes.Buffer{}
	for row := int32(0); row < img.Height; row++ {
		for column := int32(0); column < img.Width; column++ {
			pixelAsARGB, _ := img.PixelAt(row, column)
			pixelRGBA := color.RGBA{R: pixelAsARGB[1], G: pixelAsARGB[2], B: pixelAsARGB[3], A: pixelAsARGB[0]}
			pngData.Set(int(column), int(row), color.RGBA(pixelRGBA))
		}
	}
	png.Encode(&buf, pngData)
	w, err := os.Create(path);
	if err != nil {
		log.Println("Unable to write", path, err)
		return false
	}
	defer w.Close()
	w.Write(buf.Bytes())
	return true
}

