// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package resource

import (
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
)

type Resource interface {
	Path() string
	Presentation() (title string, comment string, iconUrl link.Href, profile string)
	Links(context, term string) link.List
	RelevantForSearch() bool
}

func LinkTo(res Resource, context string, rank int) link.Link {
	var path = fmt.Sprint(context, res.Path())
	var title, _, iconName, profile = res.Presentation()
	return link.MakeRanked2(link.Href(path), title, iconName, profile, rank)
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}


type ResourceRepo interface {
	GetResources() []Resource
	GetResource(path string) Resource
	Search(term string, threshold int) link.List
}
