// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package browser

import (
	"net/url"
	"strings"

	"github.com/surlykke/refude/server/applications"
	"github.com/surlykke/refude/server/lib/entity"
	"github.com/surlykke/refude/server/lib/response"
)

type Tab struct {
	entity.Base
	Id        string
	BrowserId string
	Url       string
}

func (this *Tab) DoPost(action string) response.Response {
	if app, ok := applications.AppMap.Get(this.BrowserId); ok {
		app.Run("http://refude.focustab.localhost?url=" + url.QueryEscape(this.Url))
		return response.Accepted()
	} else {
		return response.NotFound()
	}

}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
