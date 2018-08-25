// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"image"
	"bytes"
	"image/png"
	"regexp"
	"strconv"
	"strings"
	"image/color"
	"errors"
)

var stringPattern = regexp.MustCompile("^[^\"]*(\".*\")[^\"]*$")
var valuePattern = regexp.MustCompile("^\\s*(\\d+)\\s+(\\d+)\\s+(\\d+)\\s+(\\d+)(\\s+(\\d+)\\s+(\\d+))?\\s*$")
var colorPattern = regexp.MustCompile("(?i)c\\s+(\\S+)")

func extractString(line string) (string, bool) {
	if m := stringPattern.FindStringSubmatch(line); m != nil {
		if s, err := strconv.Unquote(string(m[1])); err == nil {
			return s, true
		}
	}

	return "", false
}

func extractStrings(bytes []byte) []string {
	var res []string
	for _, line := range strings.Split(string(bytes), "\n") {
		if s, ok := extractString(line); ok {
			res = append(res, s)
		}
	}
	return res
}

func str2int(str string) int {
	if i, err := strconv.Atoi(str); err != nil {
		panic(err)
	} else {
		return i
	}
}

func getValues(valuesAsString string) (width int, height int, colors int, charsPrColor int, err error) {
	if m := valuePattern.FindStringSubmatch(valuesAsString); m == nil {
		return 0, 0, 0, 0, errors.New("Malformed value section")
	} else {
		width, height, colors, charsPrColor, err = str2int(m[1]), str2int(m[2]), str2int(m[3]), str2int(m[4]), nil
		return
	}
}

func getColors(colorLines []string, charsPrColor int) (map[string]color.RGBA, error) {
	var colorMap = make(map[string]color.RGBA)
	for _, line := range colorLines {
		if len(line) < charsPrColor {
			return nil, errors.New("Malformed color line: '" + line + "'")
		} else if m := colorPattern.FindStringSubmatch(line[charsPrColor:]); m == nil {
			return nil, errors.New("Malformed color line'" + line + "' (no match)")
		} else {
			if strings.HasPrefix(m[1], "#") {
				if val, err := strconv.ParseUint("0x"+string(m[1][1:]), 0, 32); err != nil {
					return nil, errors.New("Malformed color line'" + line + "' - " + err.Error())
				} else {
					colorMap[line[:charsPrColor]] = color.RGBA{
						uint8((val & 0xFF0000) >> 16),
						uint8((val & 0xFF00) >> 8),
						uint8((val & 0xFF)),
						0xFF,
					}
				}
			} else if col, ok := NamedColors[strings.ToLower(m[1])]; !ok {
				return nil, errors.New("Unknown color name: '" + m[1] + "'")
			} else {
				colorMap[line[:charsPrColor]] = col
			}
		}
	}
	return colorMap, nil
}

func Xpm2png(data []byte) ([]byte, error) {
	if xpm := extractStrings(data); len(xpm) == 0 {
		return nil, errors.New("Value section expected")
	} else if width, heigth, colors, charsPrColor, err := getValues(xpm[0]); err != nil {
		return nil, err
	} else if len(xpm) < 1+colors+heigth {
		return nil, errors.New("Too short")
	} else if colorMap, err := getColors(xpm[1:1 + colors], charsPrColor); err != nil {
		return nil, err
	} else {
		var img = image.NewRGBA(image.Rect(0, 0, width, heigth))
		for i := 0; i < heigth; i++ {
			var line = xpm[1+colors+i]
			if len(line) != charsPrColor*width {
				return nil, errors.New("Wrong length of line: '" + line + "'")
			} else {
				for j := 0; j < width; j++ {
					var key = line[charsPrColor*j : charsPrColor*j+charsPrColor]
					if color, ok := colorMap[key]; !ok {
						return nil, errors.New("Unknown color '" + key + "' in '" + line + "'")
					} else {
						img.Set(j, i, color)
					}
				}
			}
		}

		var buf = bytes.Buffer{}
		if err = png.Encode(&buf, img); err != nil {
			return nil, err
		} else {
			return buf.Bytes(), nil
		}
	}
}
