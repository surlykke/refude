// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package browser

import (
	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/internal/lib/response"
	"github.com/surlykke/refude/internal/lib/xdg"
)

type Bookmark struct {
	entity.Base
	Id          string
	ExternalUrl string
}

func (this *Bookmark) DoPost(action string) response.Response {
	xdg.RunCmd("xdg-open", this.ExternalUrl)
	return response.Accepted()
}

// We use this for icon url
// https://t0.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&url=<bookmark url>
