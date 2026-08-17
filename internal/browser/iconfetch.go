// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package browser

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/lib/encode"
	svgsanitize "go.privatebychoice.com/pbcsvgsanitize"
)

func installIcon(iconUrl string) (osPath string, err error) {
	if bytes, err := getBytes(iconUrl); err != nil {
		return "", err
	} else {
		var fileType string
		mimetype := http.DetectContentType(bytes)
		if mimetype == "image/svg+xml" {
			if tmp, _, err := svgsanitize.Sanitize(bytes); err != nil {
				return "", err
			} else {
				bytes, fileType = tmp, "svg"
			}
		} else if mimetype == "image/png" {
			fileType = "png"
		} else if mimetype == "image/x-icon" {
			fileType = "ico"
		}

		if fileType == "" {
			return "", errors.New("No handling for mimetype " + mimetype)
		} else if osPath, err := encode.WriteCasheFile(bytes, fileType); err != nil {
			return "", err
		} else {
			return osPath, nil
		}
	}

}

func getBytes(iconUrl string) ([]byte, error) {
	const pngDataPrefix = "data:image/png;base64,"
	if strings.HasPrefix(iconUrl, pngDataPrefix) {
		return encode.Base64Decode([]byte(iconUrl[len(pngDataPrefix):]))
	} else if strings.HasPrefix(iconUrl, "https") {
		return fetch(iconUrl)
	} else {
		return nil, errors.New("No handling for url " + iconUrl)
	}
}

func fetch(url string) ([]byte, error) {
	if resp, err := getter.Get(url); err != nil {
		return nil, err
	} else if bytes, err := io.ReadAll(resp.Body); err != nil {
		return nil, err
	} else {
		return bytes, nil
	}
}

var getter = http.Client{CheckRedirect: noRedirects, Timeout: 500 * time.Millisecond}

func noRedirects(*http.Request, []*http.Request) error {
	return errors.New("No redirects allowed")
}
