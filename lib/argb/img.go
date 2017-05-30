/*
 * Copyright (c) 2017 Christian Surlykke
 *
 * This file is part of the RefudeServices project.
 * It is distributed under the GPL v2 license.
 * Please refer to the GPL2 file for a copy of the license.
 */

package argb

import (
	"image"
	"bytes"
	"image/color"
	"image/png"
	"fmt"
	"strconv"
	"net/http"
	"github.com/surlykke/RefudeServices/lib/service"
	"hash/fnv"
	"github.com/pkg/errors"
	"github.com/surlykke/RefudeServices/lib/stringlist"
)

var hashes = make(stringlist.StringList, 0)

type PNGImg struct {
	Size int
	pngData []byte
}

type PNGIcon []PNGImg


func (icon PNGIcon) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
	} else {
		var data []byte = nil

		size := 32
		if sizes, ok := r.URL.Query()["size"]; ok && len(sizes) == 1 {
			if tmp, err := strconv.Atoi(sizes[0]); err == nil {
				size = tmp
			}
		}

		for _,sizedPng := range icon {
			data = sizedPng.pngData
			if sizedPng.Size > size {
				break
			}
		}

		w.Header().Set("Content-Type", "image/png")
		w.Write(data)
	}
}

type Img struct {
	width  uint
	height uint
	/*
	 * Holds pixels in blocks of 4 bytes. Each block (from low address to high)
	 * the A,R,G and B component of the pixel
	 */
	pixels []byte
}

func (img* Img) PixelAt(row uint, column uint) ([]byte, error) {
	if column > img.width || row > img.height {
		return nil, errors.New("Out of range")
	} else {
		pos := 4*(row*img.width + column)
		return img.pixels[pos:pos + 4], nil
	}
}

type Icon []Img

/**
 * Icons retrieved from the X-server (EWMH) will come as arrays of uint. There will be first two ints giving
 * width and height, then width*height uints each holding a pixel in ARGB format (on 64bit system the 4 most
 * significant bytes are not used). After that it may repeat: again a width and height uint and then pixels and
 * sofort..
 */
func ExtractARGBIcon(uints []uint) []Img {
	res := make([]Img, 0)
	for len(uints) >= 2 {
		width := uints[0]
		height := uints[1]
		uints = uints[2:]
		if len(uints) < int(width*height) {
			break
		}
		pixels := make([]byte, 4*width*height)
		for pos := uint(0); pos < width*height; pos++ {
			pixels[4*pos] = uint8((uints[pos] & 0xFF000000) >> 24)
			pixels[4*pos + 1] = uint8((uints[pos] & 0xFF0000) >> 16)
			pixels[4*pos + 2] = uint8((uints[pos] & 0xFF00) >> 8)
			pixels[4*pos + 3] = uint8(uints[pos] & 0xFF)
		}
		res = append(res, Img{width: width, height: height, pixels: pixels})
		uints = uints[width*height:]
	}

	return res
}

func calcHash(argbIcon Icon) string {
	hash := fnv.New64a()
	for _, img := range argbIcon {
		hash.Write(img.pixels)
	}
	return fmt.Sprintf("%X", hash.Sum64())
}


func ServeAsPng(argbIcon Icon) (string, error) {
	hash := calcHash(argbIcon)
	if !hashes.Has(hash) {
		hashes = stringlist.PushBack(hashes, hash)
		pngIcon := make(PNGIcon, 0)
		for _, img := range argbIcon {
			pngData := image.NewRGBA(image.Rect(0, 0, int(img.width), int(img.height)))
			buf := bytes.Buffer{}
			for row := uint(0); row < img.height; row++ {
				for column := uint(0); column < img.width; column++ {
					pixelAsARGB,_ := img.PixelAt(row, column)
					pixelRGBA := color.RGBA{R: pixelAsARGB[1], G: pixelAsARGB[2], B: pixelAsARGB[3], A: pixelAsARGB[0] }
					pngData.Set(int(column), int(row), color.RGBA(pixelRGBA))
				}
			}
			png.Encode(&buf, pngData)
			pngIcon = append(pngIcon, PNGImg{int(img.width), buf.Bytes()})
		}
		if len(pngIcon) < 1 {
			return "", fmt.Errorf("No icons in argument")
		}
		fmt.Println("Serving: /icons/:", hashes)
		service.Map("/icons/", hashes)

		fmt.Println("Serving: /icons/" + hash)
		service.Map("/icons/" +hash, pngIcon)
	}

	return "/icons/" + hash, nil
}

