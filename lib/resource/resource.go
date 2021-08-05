package resource

import (
	"net/http"
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
}

func MakeLink(href, title, iconName string, rel relation.Relation) Link {
	return Link{
		Href:     href,
		Title:    title,
		Icon:     IconUrl(iconName),
		Relation: rel,
	}
}

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

/**
 * A resource is something that is 'linkable' and has a type.
 */
type Resource interface {
	// The resources links. Should be not empty and the first link should be the self-link
	Links() []Link
	RefudeType() string
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

type Collection []Link

func (c Collection) Links() []Link {
	return []Link(c)
}

func (c Collection) RefudeType() string {
	return "collection"
}

func (c Collection) MarshalJSON() ([]byte, error) {
	return []byte(`{}`), nil
}
