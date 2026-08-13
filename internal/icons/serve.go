// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package icons

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/respond"
	"github.com/surlykke/refude/internal/lib/utils"
)

func Run() {
	http.HandleFunc("GET /img/{iconname}", GetHandler)

	collectThemes()
	collectThemeIcons()
	collectOtherIcons()
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	var name = url.PathEscape(r.PathValue("iconname"))
	fmt.Println("icon name:", name)
	if name == "" {
		respond.UnprocessableEntity(w, errors.New("name must be given"))
	}
	var size = utils.GetUint32(r, "size", 32)

	if iconPath := FindIconPath(name, size); iconPath != "" {
		http.ServeFile(w, r, iconPath)
	} else {
		respond.NotFound(w)
	}
}

func FindIconPath(iconName string, size uint32) string {
	if namedIcon := findNamedIcon(iconName); namedIcon != nil {
		return namedIcon.GetPath(size)
	} else {
		return ""
	}
}

func findNamedIcon(iconName string) *Icon {
	if namedIcon, ok := resource.Get[*Icon](iconPrefix + iconName); ok {
		return namedIcon
	} else if lastDash := strings.LastIndex(iconName, "-"); lastDash > -1 {
		return findNamedIcon(iconName[0:lastDash]) /* By the icon naming specification, dash ('-') seperates 'levels of specificity'.
		                                               So given an icon name 'input-mouse-usb', the levels of specificy, and the names and
													   order we search will be: 'input-mouse-usb', 'input-mouse' and 'input'. */
	} else {
		return nil
	}

}

func AddFileIcon(osPath string) (string, error) {
	var name = url.PathEscape(osPath)
	if err := addIconPath(name, "", osPath, 1, 1); err != nil {
		return "", err
	} else {
		return name, nil
	}
}
