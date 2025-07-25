// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package entity

import (
	"github.com/surlykke/refude/internal/lib/icon"
	"github.com/surlykke/refude/internal/lib/response"
)

type Link struct {
	Href     string    `json:"href"`
	Title    string    `json:"title,omitempty"`
	Icon     icon.Name `json:"icon,omitempty"`
	Relation Relation  `json:"rel,omitempty"`
	Type     MediaType `json:"type,omitempty"`
}

// -------------- Serve -------------------------

type Postable interface {
	DoPost(string) response.Response
}

type Deleteable interface {
	DoDelete() response.Response
}
