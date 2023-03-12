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
	GetPath() string
	Presentation() (title string, comment string, iconName string, profile string)
	Actions() link.ActionList
	DeleteAction() (title string, ok bool)
	Links(searchTerm string) link.List
	RelevantForSearch() bool
}

type BaseResource struct {
	Path     string `json:"-"`
	Title    string `json:"-"`

	Comment  string `json:"-"`
	IconName string `json:",omitempty"`
	Profile  string `json:"-"`
}

func (br *BaseResource) GetPath() string {
	return br.Path
}

func (br *BaseResource) Presentation() (title string, comment string, iconName string, profile string) {
	return br.Title, br.Comment, br.IconName, br.Profile
}

func (br *BaseResource) Actions() link.ActionList {
	return link.ActionList{}
}

func (br *BaseResource) DeleteAction() (string, bool) {
	return "", false
}

func (br *BaseResource) Links(searchTerm string) link.List {
	return link.List{}
}

func (br *BaseResource) RelevantForSearch() bool {
	return true
}

func LinkTo(res Resource, context string, rank int) link.Link {
	var path = fmt.Sprint(context, res.GetPath())
	var title, _, iconName, profile = res.Presentation()
	return link.MakeRanked(path, title, iconName, profile, rank)
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
