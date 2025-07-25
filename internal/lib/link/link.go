// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package link

import (
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/mediatype"
	"github.com/surlykke/refude/internal/lib/relation"
	"github.com/surlykke/refude/internal/lib/response"
)

type Link struct {
	Href     string              `json:"href"`
	Title    string              `json:"title,omitempty"`
	Icon     icon.Name           `json:"icon,omitempty"`
	Relation relation.Relation   `json:"rel,omitempty"`
	Type     mediatype.MediaType `json:"type,omitempty"`
}

// -------------- Serve -------------------------

type Postable interface {
	DoPost(string) response.Response
}

type Deleteable interface {
	DoDelete() response.Response
}
