package respond

import (
	"fmt"
	"net/http"
	"strings"
)

type Relation uint8

const (
	Self Relation = 1 << iota
	DefaultAction
	Action
	Delete
	Related
	Menu
)

func (r Relation) MarshalJSON() ([]byte, error) {
	switch r {
	case Self:
		return []byte(`"self"`), nil
	case DefaultAction:
		return []byte(`"org.refude.defaultaction"`), nil
	case Action:
		return []byte(`"org.refude.action"`), nil
	case Delete:
		return []byte(`"org.refude.delete"`), nil
	case Related:
		return []byte(`"related"`), nil
	case Menu:
		return []byte(`"org.refude.menu"`), nil
	default:
		panic(fmt.Sprintf("Unknown relation: %d", r))
	}
}

type Traits []string

type Link struct {
	Href     string   `json:"href"`
	Title    string   `json:"title"`
	Icon     string   `json:"icon,omitempty"`
	Relation Relation `json:"rel"`
	Traits   Traits   `json:"traits,omitempty"`
}

type Resource struct {
	Links  []Link `json:"_links"`
	Traits Traits `json:"traits,omitempty"`
}

func MakeResource(href, title, icon string, traits ...string) Resource {
	var res = Resource{Traits: traits}
	res.AddSelfLink(href, title, icon)
	return res
}

// We will arrange for the self link to be the first, so this will perform reasonably
func (res *Resource) Self() Link {
	for _, l := range res.Links {
		if l.Relation == Self {
			return l
		}
	}
	panic("Resource has no self link")
}

func (res *Resource) GetRelatedLink() Link {
	var l = res.Self()
	l.Relation = Related
	l.Traits = res.Traits
	return l
}

func (res *Resource) AddSelfLink(href string, title string, icon string) {
	res.addLink(href, title, icon, Self)
}

func (res *Resource) AddDefaultActionLink(title string, icon string) {
	res.addLink(res.Self().Href, title, icon, DefaultAction)
}

func (res *Resource) AddActionLink(title string, icon string, actionId string) {
	var href = res.Self().Href
	var separator = "?"
	if strings.Contains(href, "?") {
		separator = "&"
	}
	res.addLink(res.Self().Href+separator+"action="+actionId, title, icon, Action)
}

func (res *Resource) AddDeleteLink(title string, icon string) {
	res.addLink(res.Self().Href, title, icon, Delete)
}

func (res *Resource) AddMenuLink(title string) {
	res.addLink(res.Self().Href+"/menu", title, "", Menu)
}

func (res *Resource) addLink(href string, title string, icon string, relation Relation) {
	res.Links = append(res.Links,
		Link{
			Href:     href,
			Title:    title,
			Icon:     icon,
			Relation: relation,
		},
	)
}

func (res *Resource) ClearNonSelfLinks() {
	res.Links = []Link{res.Self()}
}

func (res *Resource) UpdateTitle(title string) {
	for _, l := range res.Links {
		if l.Relation == Self {
			l.Title = title
			break
		}
	}
}

func (res *Resource) UpdateIcon(icon string) {
	for _, l := range res.Links {
		if l.Relation == Self {
			l.Icon = icon
			break
		}
	}
}

func (res *Resource) UpdateTraits(traits ...string) {
	res.Traits = traits
}

func (res *Resource) DoGet(this interface{}, w http.ResponseWriter, r *http.Request) {
	AsJson(w, this)
}

func (res *Resource) DoPost(w http.ResponseWriter, r *http.Request) {
	NotAllowed(w)
}

func (res *Resource) DoDelete(w http.ResponseWriter, r *http.Request) {
	NotAllowed(w)
}

type JsonResource interface {
	DoGet(this interface{}, w http.ResponseWriter, r *http.Request)
	DoPost(w http.ResponseWriter, r *http.Request)
	DoDelete(w http.ResponseWriter, r *http.Request)
}
