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
	Href     Href              `json:"href"`
	Title    string            `json:"title,omitempty"`
	Icon     Href              `json:"icon,omitempty"`
	Relation relation.Relation `json:"rel"`
	Profile  string            `json:"profile,omitempty"`
	Rank     int               `json:"-"` // Used when searching
}

func Make(href, title, iconName string, rel relation.Relation) Link {
	return Link{
		Href:     Href(href),
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	}
}

func MakeRanked(href, title, iconName string, profile string, rank int) Link {
	return MakeRanked2(Href(href), title, IconUrl(iconName), profile, rank)
}

func MakeRanked2(href Href, title string, icon Href, profile string, rank int) Link {
	return Link{
		Href:     Href(href),
		Title:    title,
		Icon:     icon,
		Relation: relation.Related,
		Profile:  profile,
		Rank:     rank,
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

func (list List) Add(href, title, iconName string, rel relation.Relation) List {
	return append(list, Link{
		Href:     Href(href),
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	})
}

func (list List) SelfLink() (Link, bool) {
	for _, lnk := range list {
		if lnk.Relation == relation.Self {
			return lnk, true
		}
	}
	return Link{}, false
}

// ---------- Implement sort.Sort ------------------------------------
func (list List) Len() int { return len(list) }

func (list List) Less(i int, j int) bool {
	if list[i].Rank == list[j].Rank {
		return list[i].Href < list[j].Href // Not that Href is particularly relevant, sortingwise. Just to have a reproducible order
	} else {
		return list[i].Rank < list[j].Rank
	}
}

func (list List) Swap(i int, j int) { list[i], list[j] = list[j], list[i] }

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
