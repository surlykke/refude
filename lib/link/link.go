package link

import (
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type Link struct {
	Href       string            `json:"href"`
	Title      string            `json:"title"`
	Icon       string            `json:"icon,omitempty"`
	Relation   relation.Relation `json:"rel"`
	RefudeType string            `json:"refudeType,omitempty"`
	Rank       int               `json:"-"` // Used when searching
}

func Make(href, title, iconName string, rel relation.Relation) Link {
	return Link{
		Href:     href,
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	}
}

func MakeRanked(href, title, iconName string, refudeType string, rank int) Link {
	return Link{
		Href:       href,
		Title:      title,
		Icon:       IconUrl(iconName),
		Relation:   relation.Related,
		RefudeType: refudeType,
		Rank:       rank,
	}
}

type List []Link

func MakeList(href, title, iconName string) List {
	return List{{
		Href:     href,
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: relation.Self,
	}}
}

func (ll List) Add(href, title, iconName string, rel relation.Relation) List {
	return append(ll, Link{
		Href:     href,
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	})
}

func (ll List) Add2(href, title, iconName string, refudeType string, rank int) List {
	return append(ll, Link{
		Href:       href,
		Title:      title,
		Icon:       IconUrl(iconName),
		Relation:   relation.Related,
		RefudeType: refudeType,
		Rank:       rank,
	})
}

// ---------- Implement sort.Sort ------------------------------------
func (ll List) Len() int { return len(ll) }

func (ll List) Less(i int, j int) bool {
	if ll[i].Rank == ll[j].Rank {
		return ll[i].Href < ll[j].Href // Not that Href is particularly relevant, sortingwise. Just to have a reproducible order
	} else {
		return ll[i].Rank < ll[j].Rank
	}
}

func (ll List) Swap(i int, j int) { ll[i], ll[j] = ll[j], ll[i] }

// --------------------------------------------------------------------

type Collection List

func (c Collection) Links() List {
	return List(c)
}

func (c Collection) RefudeType() string {
	return "collection"
}

func (c Collection) MarshalJSON() ([]byte, error) {
	return []byte(`{}`), nil
}

// --------------------------------------------------------------------

func IconUrl(name string) string {
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
		return "/icon?name=" + url.QueryEscape(name)
	} else {
		return ""
	}
}
