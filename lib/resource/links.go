// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

const LinksMediaType MediaType = "application/vnd.org.refude.links+json"

type Links map[MediaType][]StandardizedPath

func (l *Links) GetSelf() StandardizedPath {
	return "/links"
}

func (l *Links) GetMt() MediaType {
	return LinksMediaType
}



