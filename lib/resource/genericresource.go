package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type Executer func()

type ResourceAction struct {
	Description string
	IconName    string
	Executer    Executer `json:"-"`
}

type GenericResource struct {
	Self            StandardizedPath          `json:"_self"` // Convenience - is also contained in Links
	Links           []Link                    `json:"_links"`
	Mt              MediaType                 `json:"-"`
	ResourceActions map[string]ResourceAction `json:"_actions,omitempty"`
}

func MakeGenericResource(SelfLink StandardizedPath, mt MediaType) GenericResource {
	return GenericResource{
		Self:            SelfLink,
		Links:           []Link{{Href: SelfLink, Rel: Self}},
		Mt:              mt,
		ResourceActions: make(map[string]ResourceAction),
	}
}

func (gr *GenericResource) GetSelf() StandardizedPath {
	for _, link := range gr.Links {
		if link.Rel == Self {
			return link.Href
		}
	}

	panic("Resource has no self link")
}

func (gr *GenericResource) GetMt() MediaType {
	return gr.Mt
}

func (gr *GenericResource) LinkTo(target StandardizedPath, relation Relation) {
	gr.Links = append(gr.Links, Link{Href: target, Rel: relation})
}

func (gr *GenericResource) POST(w http.ResponseWriter, r *http.Request) {
	var actionId = requests.GetSingleQueryParameter(r, "action", "default")
	if action, ok := gr.ResourceActions[actionId]; ok {
		action.Executer()
		w.WriteHeader(http.StatusAccepted)
	} else {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
}

func (gr *GenericResource) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (gr *GenericResource) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}
