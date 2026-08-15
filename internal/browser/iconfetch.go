package browser

import (
	"encoding/base64"
	"io"
	"net/http"
	"strings"

	svgsanitize "go.privatebychoice.com/pbcsvgsanitize"

)

func installIcon(iconUrl string) string {
	if strings.HasPrefix(iconUrl, "data:image") {
		return installImage(iconUrl)
	} else if strings.HasPrefix(iconUrl, "https") {
		installRemote(iconUrl)
	}
	return ""
}

func installImage(img string) string {
	if strings.HasPrefix(img, "data:image/png;base64,") {
		var imgAsBase64 = []byte(strings.Replace(img, "data:image/png;base64,", "", 1))
		var decoded = make([]byte, base64.StdEncoding.DecodedLen(len(imgAsBase64)))
		if _, err := base64.StdEncoding.Decode(decoded, imgAsBase64); err != nil {
			return ""
		}
		return installPng(decoded)
	} else {
		// TODO: Other formats?
		return ""
	}
}

func installRemote(url string) string {
	if r, err := getter.Get(url); err != nil {
		return ""
	} else if bytes, err := io.ReadAll(r.Body); err != nil  {
		return ""
	} else if mimetype := r.Header.Get("content-type"); mimetype == "image/svg+xml" {
		if sanitized, _, err := svgsanitize.Sanitize(bytes); err != nil {
			return ""
		} else {
			

	}
	return ""
}

func installPng(png []byte) string {
	return ""
}

func installSvg(svg []byte) string {
	return ""
}

var getter = http.Client{CheckRedirect:  func(*http.Request, []*http.Request) error { return errors.New("No redirects allowed")}}
