// Copyright (c) 2017,2018 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package lib

import (
	"strings"
)

type MediaType string

func MediaTypeMatch(sought, actual MediaType) bool {
	if sought == "" {
		return true
	} else if sought == actual {
		return true
	} else {
		// If a link has, eg., mediatype 'application/vnd.org.refude.desktopapplication+json'
		// and 'application/json' is sought we consider that a match
		var slashPos = strings.Index(string(actual), "/")
		var plusPos = strings.Index(string(actual), "+")
		if 0 < slashPos && slashPos < plusPos {
			var generelType = actual[:slashPos + 1] + actual[plusPos + 1:]
			if sought == generelType {
				return true
			}
		}
	}
	return false
}





