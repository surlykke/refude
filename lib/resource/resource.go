// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"golang.org/x/exp/constraints"
)

type Resource[ID constraints.Ordered] interface {
	Id() ID
	Presentation() (title string, comment string, iconUrl link.Href, profile string)
	Links(self, term string) link.List
}

func LinkTo[ID constraints.Ordered](res Resource[ID], context string, rank int) link.Link {
	var path = fmt.Sprint(context, res.Id())
	var title, _, iconName, profile = res.Presentation()
	return link.MakeRanked2(link.Href(path), title, iconName, profile, rank)
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Wrapper struct { // Maybe generic?
	Self    link.Href   `json:"self"`
	Links   link.List   `json:"links"`
	Title   string      `json:"title"`
	Comment string      `json:"comment,omitempty"`
	Icon    link.Href   `json:"icon,omitempty"`
	Profile string      `json:"profile"`
	Data    interface{} `json:"data"`
}

func MakeWrapper[ID constraints.Ordered, T Resource[ID]](self string, res T, linkSearchTerm string) Wrapper {
	var wrapper = Wrapper{}
	wrapper.Self = link.Href(self)
	wrapper.Links = res.Links(self, linkSearchTerm)
	wrapper.Data = res
	wrapper.Title, wrapper.Comment, wrapper.Icon, wrapper.Profile = res.Presentation()
	return wrapper
}
