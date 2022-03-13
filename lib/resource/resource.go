// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
)

type Resource interface {
	Self() string
	Presentation() (title string, comment string, iconUrl link.Href, profile string)
	Links(term string) link.List
}


func LinkTo(res Resource, rank int) link.Link {
	var path = res.Self()
	var title, _, iconName, profile = res.Presentation()
	return link.MakeRanked2(link.Href(path), title, iconName, profile, rank)
}


type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Wrapper struct {
	Self    link.Href   `json:"self"`
	Links   link.List   `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    link.Href   `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func MakeWrapper(res Resource, linkSearchTerm string) Wrapper {
	var wrapper = Wrapper{}
	wrapper.Self, wrapper.Links, wrapper.Data = link.Href(res.Self()), res.Links(linkSearchTerm), res
	wrapper.Title, wrapper.Comment, wrapper.Icon, wrapper.Profile = res.Presentation()
	return wrapper
}


