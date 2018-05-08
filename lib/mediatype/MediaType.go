package mediatype

import (
	"encoding/json"
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




func ToJSon(res interface{}) []byte {
	if bytes, err := json.Marshal(res); err != nil {
		panic("Could not json-marshal")
	} else {
		return bytes
	}
}

