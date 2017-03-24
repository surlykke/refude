/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package main

import (
	"image"
	"bytes"
	"image/color"
	"image/png"
	"net/http"
	"fmt"
	"strconv"
)

type SizedPng struct {
	Size int
	pngData []byte
}

type Icon struct {
	hash uint64
	pngs []SizedPng
}

func (icon Icon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var bytes []byte = nil

		size := 32
		if sizes, ok := r.URL.Query()["size"]; ok && len(sizes) == 1 {
		   if tmp, err := strconv.Atoi(sizes[0]); err == nil {
			   size = tmp
		   }
		}

		for _,sizedPng := range icon.pngs {
			bytes = sizedPng.pngData
			if sizedPng.Size > size {
				break
			}
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(bytes)
	}
}

func MakeIcon(hash uint64, iconArr []uint) (Icon, error) {
	res := Icon{hash, make([]SizedPng, 0)}

	for pos := 0; pos < len(iconArr); {
		width := int(iconArr[pos])
		height := int(iconArr[pos + 1])
		img := image.NewRGBA(image.Rect(0, 0, width, height))
		buf := bytes.Buffer{}
		for row :=0 ; row < height; row++ {
			for column := 0; column < width; column++ {
				pixelARGB := iconArr[pos + 2 + row*width + column]
				pixelRGBA := color.RGBA{uint8((pixelARGB & 0xFF0000) >> 16),
				                        uint8((pixelARGB & 0xFF00) >> 8),
				                        uint8(pixelARGB & 0xFF),
				                        uint8((pixelARGB & 0xFF000000) >> 24)}
				img.Set(column, row, color.RGBA(pixelRGBA))
			}
		}
		png.Encode(&buf, img)
		res.pngs = append(res.pngs, SizedPng{width, buf.Bytes()})
		pos = pos + 2 + width*height
	}

	if len(res.pngs) < 1 {
		return res, fmt.Errorf("No icons in argument")
	} else {
		return res, nil
	}
}

