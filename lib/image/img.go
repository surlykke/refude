// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package image

import (
	"bytes"
	"crypto/sha1"
	"image"
	"image/color"
	"image/png"

	"github.com/pkg/errors"
)

type ARGBImage struct {
	Width  uint32
	Height uint32
	/**
	 * Holds pixels in blocks of 4 bytes. Each block (from low address to high)
	 * the R,G,B and A component of the pixel
	 * Pixels are arranged left-to-right, top to bottom, so Pixels[0:4] is the leftmost pixel of the top row
	 * Pixel[4*(width-1):4*width] is the right-most pixel of the top row, Pixel[4*width:4*width+4] leftmost of second row etc.
	 */
	Pixels []byte
}

func (a *ARGBImage) PixelAt(row uint32, column uint32) ([]byte, error) {
	if column < 0 || column > a.Width || row < 0 || row > a.Height {
		return nil, errors.New("Out of range")
	} else {
		pos := 4 * (row*a.Width + column)
		return a.Pixels[pos : pos+4], nil
	}
}

func (a *ARGBImage) AsPng() ([]byte, error) {
	pngData := image.NewRGBA(image.Rect(0, 0, int(a.Width), int(a.Height)))
	buf := &bytes.Buffer{}
	for row := uint32(0); row < a.Height; row++ {
		for column := uint32(0); column < a.Width; column++ {
			pixelAsARGB, _ := a.PixelAt(row, column)
			pixelRGBA := color.RGBA{R: pixelAsARGB[1], G: pixelAsARGB[2], B: pixelAsARGB[3], A: pixelAsARGB[0]}
			pngData.Set(int(column), int(row), color.RGBA(pixelRGBA))
		}
	}
	if err := png.Encode(buf, pngData); err != nil {
		return nil, err
	} else {
		return buf.Bytes(), nil
	}
}

type ARGBIcon struct {
	Name   string
	Images []ARGBImage
}

func MakeIconWithHashAsName(imagelist []ARGBImage) ARGBIcon {
	var hasher = sha1.New()
	for _, image := range imagelist {
		hasher.Write(image.Pixels)
	}
	return ARGBIcon{string(hasher.Sum(nil)), imagelist}
}
