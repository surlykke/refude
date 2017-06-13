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
	"github.com/surlykke/RefudeServices/lib/resource"
)

type PNGImg struct {
	Size int
	pngData []byte
}

type PNGIcon []PNGImg


func PNGIconGET(this *resource.Resource, w http.ResponseWriter, r *http.Request) {
    icon := this.Data.(PNGIcon)
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

type Img struct {
	Width  int32
	Height int32
	/*
	 * Holds pixels in blocks of 4 bytes. Each block (from low address to high)
	 * the A,R,G and B component of the pixel
	 */
	Pixels []byte
}

func (img* Img) PixelAt(row int32, column int32) ([]byte, error) {
	if column < 0 || column > img.Width || row < 0 || row > img.Height {
		return nil, errors.New("Out of range")
	} else {
		pos := 4*(row*img.Width + column)
		return img.Pixels[pos:pos + 4], nil
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
		width := int32(uints[0])
		height := int32(uints[1])
		uints = uints[2:]
		if len(uints) < int(width*height) {
			break
		}
		pixels := make([]byte, 4*width*height)
		for pos := int32(0); pos < width*height; pos++ {
			pixels[4*pos] = uint8((uints[pos] & 0xFF000000) >> 24)
			pixels[4*pos + 1] = uint8((uints[pos] & 0xFF0000) >> 16)
			pixels[4*pos + 2] = uint8((uints[pos] & 0xFF00) >> 8)
			pixels[4*pos + 3] = uint8(uints[pos] & 0xFF)
		}
		res = append(res, Img{Width: width, Height: height, Pixels: pixels})
		uints = uints[width*height:]
	}

	return res
}

func ServeAsPng(argbIcon Icon) (string, error) {
	hash := fnv.New64a()
	for _, img := range argbIcon {
		hash.Write(img.Pixels)
	}
	path := fmt.Sprintf("/icons/%X", hash.Sum64())

	if !service.Has(path) {
		pngIcon := make(PNGIcon, 0)
		for _, img := range argbIcon {
			pngData := image.NewRGBA(image.Rect(0, 0, int(img.Width), int(img.Height)))
			buf := bytes.Buffer{}
			for row := int32(0); row < img.Height; row++ {
				for column := int32(0); column < img.Width; column++ {
					pixelAsARGB,_ := img.PixelAt(row, column)
					pixelRGBA := color.RGBA{R: pixelAsARGB[1], G: pixelAsARGB[2], B: pixelAsARGB[3], A: pixelAsARGB[0] }
					pngData.Set(int(column), int(row), color.RGBA(pixelRGBA))
				}
			}
			png.Encode(&buf, pngData)
			pngIcon = append(pngIcon, PNGImg{int(img.Width), buf.Bytes()})
		}
		if len(pngIcon) < 1 {
			return "", fmt.Errorf("No icons in argument")
		}

		service.Map(path, &resource.Resource{Data: pngIcon, GET: PNGIconGET})
	}

	return path, nil
}

