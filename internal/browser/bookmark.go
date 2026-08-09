// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the file GPL2 for a copy of the license.
package browser

import (
	"github.com/surlykke/refude/internal/lib/resource"
	"github.com/surlykke/refude/internal/lib/xdg"
)

type Bookmark struct {
	resource.Base
	Id          string
	ExternalUrl string
}

func (this *Bookmark) DoPost(action string) (bool, error) {
	xdg.RunCmd("xdg-open", this.ExternalUrl)
	return true, nil
}

// We use this for icon url
// https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=<bookmark url>
