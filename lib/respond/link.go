package respond

import "net/http"

type Relation string

const (
	None    Relation = ""
	Self             = "self"
	Related          = "related"
	Action           = "org.refude.relations.Action"
)

type Link struct {
	Href    string            `json:"href"`
	Title   string            `json:"title"`
	Icon    string            `json:"icon,omitempty"`
	Rel     Relation          `json:"rel,omitempty"`
	Profile string            `json:"profile,omitempty"`
	Meta    map[string]string `json:"meta,omitempty"`
	Rank    int               `json:"-"`
}

type Links []Link

/**
 * For use through embedding. Assumes the first link in ll is the containing resources self link
 */
func (ll Links) Link() Link {
	var copy = ll[0]
	copy.Rel = Related
	return copy
}

func (ll Links) Add(href string, title string, icon string, rel Relation, profile string, meta map[string]string) Links {
	return append(ll, Link{
		Href:    href,
		Title:   title,
		Icon:    icon,
		Rel:     rel,
		Profile: profile,
	})
}

func (ll Links) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		AsJson(w, ll)
	} else {
		NotAllowed(w)
	}
}

// -- implement sort.Interface
func (sfl Links) Len() int { return len(sfl) }

func (sfl Links) Less(i int, j int) bool {
	if sfl[i].Rank == sfl[j].Rank {
		return sfl[i].Href < sfl[j].Href // Not that Href is particularly relevant, sortingwise. Just to have a reproducible order
	} else {
		return sfl[i].Rank < sfl[j].Rank
	}
}

func (sfl Links) Swap(i int, j int) { sfl[i], sfl[j] = sfl[j], sfl[i] }

// ---
