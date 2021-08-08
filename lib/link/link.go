package link

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/xdg"
)

type Href string

var httpLocalHost7838 = []byte("http://localhost:7938")
var controlEscape = [][]byte{
	[]byte(`\u0000`), []byte(`\u0001`), []byte(`\u0002`), []byte(`\u0003`), []byte(`\u0004`), []byte(`\u0005`), []byte(`\u0006`), []byte(`\u0007`),
	[]byte(`\u0008`), []byte(`\u0009`), []byte(`\u000A`), []byte(`\u000B`), []byte(`\u000C`), []byte(`\u000D`), []byte(`\u000E`), []byte(`\u000F`),
	[]byte(`\u0010`), []byte(`\u0011`), []byte(`\u0012`), []byte(`\u0013`), []byte(`\u0014`), []byte(`\u0015`), []byte(`\u0016`), []byte(`\u0017`),
	[]byte(`\u0018`), []byte(`\u0019`), []byte(`\u001A`), []byte(`\u001B`), []byte(`\u001C`), []byte(`\u001D`), []byte(`\u001E`), []byte(`\u001F`),
}
var quoteEscape = []byte(`\"`)
var backslashEscape = []byte(`\\`)

func (href Href) MarshalJSON() ([]byte, error) {
	var buf = &bytes.Buffer{}
	buf.WriteByte('"')
	buf.Write(httpLocalHost7838)
	for _, b := range []byte(href) {
		if b <= 0x1F {
			buf.Write(controlEscape[b])
		} else if b == '\\' {
			buf.Write(backslashEscape)
		} else if b == '"' {
			buf.Write(quoteEscape)
		} else {
			buf.WriteByte(b)
		}
	}
	buf.WriteByte('"')
	return buf.Bytes(), nil
}

type Link struct {
	Href       Href              `json:"href"`
	Title      string            `json:"title"`
	Icon       Href              `json:"icon,omitempty"`
	Relation   relation.Relation `json:"rel"`
	RefudeType string            `json:"refudeType,omitempty"`
	Rank       int               `json:"-"` // Used when searching
}

func Make(href, title, iconName string, rel relation.Relation) Link {
	return Link{
		Href:     Href(href),
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	}
}

func MakeRanked(href, title, iconName string, refudeType string, rank int) Link {
	return Link{
		Href:       Href(href),
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
		Href:     Href(href),
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: relation.Self,
	}}
}

func (ll List) Add(href, title, iconName string, rel relation.Relation) List {
	return append(ll, Link{
		Href:     Href(href),
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	})
}

func (ll List) Add2(href, title, iconName string, refudeType string, rank int) List {
	return append(ll, Link{
		Href:       Href(href),
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

func IconUrl(name string) Href {
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
		return Href("/icon?name=" + url.QueryEscape(name))
	} else {
		return ""
	}
}
