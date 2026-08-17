// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
//
package browser

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/surlykke/refude/internal/lib/utils"
	"github.com/surlykke/refude/internal/rawfile"
	svgsanitize "go.privatebychoice.com/pbcsvgsanitize"
)

func installIcon(iconUrl string) string {
	if iconUrl == "" {
		return ""
	} else if strings.HasPrefix(iconUrl, "data:image/png;base64,") {
		fmt.Println("Installing data:image...")
		if decodedBytes := utils.Base64Decode([]byte(iconUrl[22:])); decodedBytes == nil {
			return ""
		} else {
			return rawfile.InstallAndRegisterBytes(decodedBytes, "png")
		}
	} else if strings.HasPrefix(iconUrl, "https") {
		fmt.Println("Installing:", iconUrl)
		if bytes, err := fetch(iconUrl); err != nil {
			log.Println(err)
			return ""
		} else if len(bytes) == 0 {
			log.Println("Empty body received")
			return ""
		} else {
			var mimetype = http.DetectContentType(bytes)  // Slightly more reliable than looking at 'content-type' 
			switch mimetype {
			case "image/svg+xml": 
				if sanitized, _, err := svgsanitize.Sanitize(bytes); err != nil {
					return ""
				} else {
					return rawfile.InstallAndRegisterBytes(sanitized, "svg")
				}
			case "image/png":
				return rawfile.InstallAndRegisterBytes(bytes, "png")
			case "image/x-icon":
				return rawfile.InstallAndRegisterBytes(bytes, "ico")
			default:
				log.Println("Can't handle mimetype:", mimetype)
				return ""
			}
		}
	} else {
		log.Println("installIcon not handling url", iconUrl)
		return ""
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

var getter = http.Client{CheckRedirect: noRedirects, Timeout: 200 * time.Millisecond}

func noRedirects(*http.Request, []*http.Request) error {
	return errors.New("No redirects allowed")
}
