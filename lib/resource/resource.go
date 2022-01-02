package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/relation"
	"github.com/surlykke/RefudeServices/lib/respond"
)

type Data interface {
	Links(path string) link.List
}

type Resource struct {
	Links   link.List `json:"_links"`
	Path    string    `json:"-"`
	Title   string    `json:"title"`
	Comment string    `json:"comment,omitempty"`
	Icon    link.Href `json:"icon,omitempty"`
	Profile string    `json:"profile"`
	Data    Data      `json:"data"`
}

func MakeResource(path, title, comment, iconName, profile string, data Data) *Resource {
	return &Resource{
		Links:   append(link.List{link.Make(path, "", "", relation.Self)}, data.Links(path)...),
		Path:    path,
		Title:   title,
		Comment: comment,
		Icon:    link.IconUrl(iconName),
		Profile: profile,
		Data:    data,
	}
}

func (res *Resource) MakeRankedLink(rank int) link.Link {
	return link.MakeRanked2(res.Links[0].Href, res.Title, res.Icon, res.Profile, rank)
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}

func (res *Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		respond.AsJson(w, res)
		return
	case "POST":
		if postable, ok := res.Data.(Postable); ok {
			postable.DoPost(w, r)
			return
		}
	case "DELETE":
		if deleteable, ok := res.Data.(Deleteable); ok {
			deleteable.DoDelete(w, r)
			return
		}
	}
	respond.NotAllowed(w)

}

type dataSlice []*Resource

func (ds dataSlice) Links(path string) link.List {
	return link.List{}
}
