// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package browser

import (
	"strings"

	"github.com/surlykke/refude/internal/lib/entity"
	"github.com/surlykke/refude/pkg/bind"
)

type Tab struct {
	entity.Base
	Id        string
	BrowserId string
	Url       string
}

func (this *Tab) DoPost(action string) bind.Response {
	browserCommands.Publish(browserCommand{BrowserId: this.BrowserId, TabId: this.Id, Cmd: "focus"})
	return bind.Accepted()
}

func (this *Tab) OmitFromSearch() bool {
	return strings.HasPrefix(this.Url, "http://localhost:7938/desktop")
}
