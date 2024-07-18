// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package link

import (
	"net/url"
	"slices"
	"strings"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

var httpLocalHost7838 = []byte("http://localhost:7938")
var controlEscape = [][]byte{
	[]byte(`\u0000`), []byte(`\u0001`), []byte(`\u0002`), []byte(`\u0003`), []byte(`\u0004`), []byte(`\u0005`), []byte(`\u0006`), []byte(`\u0007`),
	[]byte(`\u0008`), []byte(`\u0009`), []byte(`\u000A`), []byte(`\u000B`), []byte(`\u000C`), []byte(`\u000D`), []byte(`\u000E`), []byte(`\u000F`),
	[]byte(`\u0010`), []byte(`\u0011`), []byte(`\u0012`), []byte(`\u0013`), []byte(`\u0014`), []byte(`\u0015`), []byte(`\u0016`), []byte(`\u0017`),
	[]byte(`\u0018`), []byte(`\u0019`), []byte(`\u001A`), []byte(`\u001B`), []byte(`\u001C`), []byte(`\u001D`), []byte(`\u001E`), []byte(`\u001F`),
}
var quoteEscape = []byte(`\"`)
var backslashEscape = []byte(`\\`)

type Link struct {
	Href     string            `json:"href"`
	Title    string            `json:"title,omitempty"`
	IconUrl  string            `json:"icon,omitempty"`
	Relation relation.Relation `json:"rel,omitempty"`
	Profile  string            `json:"profile,omitempty"`
}

type LinkList []Link

func (ll LinkList) Add(href, title, iconUrl string, relation relation.Relation, profile string) LinkList {
	var res = slices.Clone(ll)
	return append(res, Link{Href: normalizeHref(href), Title: title, IconUrl: iconUrl, Relation: relation, Profile: profile})
}

// Ensure only one link with given relation in list
func (ll LinkList) Set(href, title, iconUrl string, relation relation.Relation, profile string) LinkList {
	var res = make(LinkList, len(ll)+1, len(ll)+1)
	var pos = 0
	for i := 0; i < len(ll); i++ {
		if ll[i].Relation != relation {
			res[pos] = ll[i]
			pos++
		}
	}
	res[pos] = Link{Href: normalizeHref(href), Title: title, IconUrl: iconUrl, Relation: relation, Profile: profile}
	res = res[0 : pos+1]
	return res
}

// Gets first found link with given relation
func (ll LinkList) Get(relation relation.Relation) (Link, bool) {
	for _, l := range ll {
		if l.Relation == relation {
			return l, true
		}
	}
	return Link{}, false
}

func (ll LinkList) GetHref(relation relation.Relation) string {
	if l, ok := ll.Get(relation); ok {
		return l.Href
	} else {
		return ""
	}
}

func normalizeHref(href string) string {
	if strings.HasPrefix(href, "/") {
		return "http://localhost:7938" + href
	} else {
		return href
	}
}

// --------------------------------------------------------------------

func IconUrlFromName(name string) string {
	if strings.Index(name, "/") > -1 {
		// So its a path..
		if strings.HasPrefix(name, "file:///") {
			name = name[7:]
		} else if strings.HasPrefix(name, "file://") {
			name = xdg.Home + "/" + name[7:]
		} else if !strings.HasPrefix(name, "/") {
			name = xdg.Home + "/" + name
		}

		// Maybe: Check that path points to iconfile..
	}
	if name != "" {
		return "http://localhost:7938/icon?name=" + url.QueryEscape(name)
	} else {
		return ""
	}
}
