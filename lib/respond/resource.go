package respond

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type Link struct {
	Href    string  `json:"href"`
	Title   string  `json:"title"`
	Icon    string  `json:"icon,omitempty"`
	Options Options `json:"options,omitempty"`
	Traits  Traits  `json:"traits,omitempty"`

	Rank int `json:"-"`
}

func MakeLink(href, title, icon string) Link {
	return Link{
		Href:   href,
		Title:  title,
		Icon:   icon,
		Traits: nil,
	}
}

// -- implement sort.Interface
type LinkList []Link

func (ll LinkList) Len() int { return len(ll) }

func (ll LinkList) Less(i int, j int) bool {
	if ll[i].Rank == ll[j].Rank {
		return ll[i].Href < ll[j].Href // Not that Href is particulally relevant, sortingwise. Just to have a reproducible order
	} else {
		return ll[i].Rank < ll[j].Rank
	}
}

func (ll LinkList) Swap(i int, j int) { ll[i], ll[j] = ll[j], ll[i] }

type Actor func(*http.Request) error

type Action struct {
	ActionId string `json:"actionId"`
	Title    string `json:"title"`
	Icon     string `json:"icon,omitempty"`
	actor    Actor
}

func MakeAction(actionId, title, icon string, actor Actor) Action {
	return Action{
		ActionId: actionId,
		Title:    title,
		Icon:     icon,
		actor:    actor,
	}
}

func (a *Action) String() string {
	return "{ActionId:'" + a.ActionId + "', Title:'" + a.Title + "', Icon:'" + a.Icon + "'}"
}

type Options struct {
	POST   []Action `json:",omitempty"`
	DELETE *Action  `json:",omitempty"`
}

func (o *Options) String() string {
	var post = make([]string, 0, len(o.POST))
	for _, act := range o.POST {
		post = append(post, act.String())
	}
	var del = "null"
	if o.DELETE != nil {
		del = o.DELETE.String()
	}
	return "{POST:[" + strings.Join(post, ",") + "],DELETE:" + del + "}"
}

type Traits []string

func (tt Traits) String() string {
	var result = "["
	for i, t := range tt {
		if i > 0 {
			result += ","
		}
		result += ("'" + t + "'")
	}
	result += "]"
	return result
}

type Resource struct {
	Self    Link        `json:"_self"`
	Related []Link      `json:"_related"`
	Owner   interface{} `json:"-"`
}

// Must be followed by a call to SetSelf, before serving
func MakeResource(href, title, icon string, owner interface{}, traits ...string) Resource {
	return Resource{
		Self: Link{
			Href:   href,
			Title:  title,
			Icon:   icon,
			Traits: traits,
		},
		Related: []Link{},
		Owner:   owner,
	}
}

func MakeRelatedCollection(href, title string, related []Link) *Resource {
	var res = MakeResource(href, title, "", nil, "list")
	res.Related = related
	return &res
}

func (res *Resource) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		if res.Owner != nil {
			AsJson(w, res.Owner)
		} else {
			AsJson(w, res)
		}
	} else if r.Method == "OPTIONS" {
		var allow = "OPTIONS, GET"
		if len(res.Self.Options.POST) > 0 {
			allow += ", POST"
		}
		if res.Self.Options.DELETE != nil {
			allow += ", DELETE"
		}
		w.Header().Add("Allow", allow)
		if allow != "OPTIONS, GET" {
			AsJson(w, res.Self.Options)
		} else {
			Ok(w)
		}
	} else if r.Method == "POST" && res.Self.Options.POST != nil {
		var actionId = url.QueryEscape(requests.GetSingleQueryParameter(r, "actionId", ""))
		fmt.Println("actionId:", actionId)
		for _, actionLink := range res.Self.Options.POST {
			if actionId == actionLink.ActionId {
				fmt.Println("found")
				runActor(w, r, actionLink.actor)
				return
			}
		}
		UnprocessableEntity(w, fmt.Errorf("Missing or unknown actionId"))

	} else if r.Method == "DELETE" && res.Self.Options.DELETE != nil {
		runActor(w, r, res.Self.Options.DELETE.actor)
	} else {
		NotAllowed(w)
	}
}

func runActor(w http.ResponseWriter, r *http.Request, actor Actor) {
	if err := actor(r); err != nil {
		ServerError(w, err)
	} else {
		Accepted(w)
	}
}

func (res *Resource) SetSelf(href string, title string, icon string) {
	res.Self = Link{
		Href:  href,
		Title: title,
		Icon:  icon,
	}
}

func (res *Resource) AddLink(link ...Link) {
	res.Related = append(res.Related, link...)
}

func (res *Resource) AddAction(action Action) {
	res.Self.Options.POST = append(res.Self.Options.POST, action)
}

func (res *Resource) ClearActions() {
	res.Self.Options.POST = nil
}

func (res *Resource) GetRelatedLink(rank int) Link {
	var linkToThis = res.Self
	linkToThis.Rank = rank
	return linkToThis
}

func stringHelper(b *strings.Builder, ss ...string) {
	for _, s := range ss {
		b.WriteString(s)
	}
}
