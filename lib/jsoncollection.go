// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

const LinksMediaType MediaType = "application/vnd.org.refude.links+json"

type JsonCollection interface {
	GetResource(path StandardizedPath) *JsonResource
	GetAll() []*JsonResource
}

type Links map[MediaType][]StandardizedPath

func (Links) GetSelf() StandardizedPath {
	return "/links"
}

func (Links) GetMt() MediaType {
	return LinksMediaType
}



