package entity

import (
	"github.com/surlykke/RefudeServices/lib/icon"
	"github.com/surlykke/RefudeServices/lib/link"
	"github.com/surlykke/RefudeServices/lib/mediatype"
	"github.com/surlykke/RefudeServices/lib/relation"
)

type Base struct {
	path      string
	Title     string
	Icon      icon.Name
	MediaType mediatype.MediaType
	Links     []link.Link
	Keywords  []string `json:"keywords"`
	Actions   []Action `json:"actions"`
}

type Short struct {
	Title     string
	Icon      icon.Name
	Keywords  []string
	Links     []link.Link
	MediaType mediatype.MediaType
}

type Action struct {
	Id   string
	Name string
	Icon icon.Name
}

func MakeBase(title string, icon icon.Name, mediatype mediatype.MediaType, keywords ...string) *Base {
	return &Base{
		Title:     title,
		Icon:      icon,
		MediaType: mediatype,
		Keywords:  keywords,
	}
}

func (this *Base) OmitFromSearch() bool {
	return false
}

func (this *Base) GetShort() Short {
	return Short{
		Title:     this.Title,
		Icon:      this.Icon,
		Keywords:  this.Keywords,
		Links:     this.Links,
		MediaType: this.MediaType,
	}
}

// ------------ Don't call after published ------------------

func (this *Base) GetPath() string {
	return this.path
}

func (this *Base) SetPath(path string) {
	this.path = path
	this.Links = make([]link.Link, 1+len(this.Actions), 1+len(this.Actions))
	this.Links[0] = link.Link{Href: this.path, Title: this.Title, Icon: this.Icon, Relation: relation.Self}
	for i, action := range this.Actions {
		var actionParam string
		if action.Id != "" {
			actionParam = "?action=" + action.Id
		}
		this.Links[i+1] = link.Link{Href: this.path + actionParam, Title: action.Name, Relation: relation.Action}
	}
}

func (this *Base) AddAction(id string, name string, iconUrl icon.Name) {
	this.Actions = append(this.Actions, Action{Id: id, Name: name})
}

func (this *Base) ActionLinks() []link.Link {
	var actionLinks = make([]link.Link, 0, len(this.Links))
	for _, l := range this.Links {
		if l.Relation == relation.Action || l.Relation == relation.Delete {
			actionLinks = append(actionLinks, l)
		}
	}
	return actionLinks
}

/*func (this *ResourceData) AddDeleteAction(actionId string, title string, comment string, iconName icon.Name) {
	this.Links = append(this.Links, Link{Href: href.Of(this.Path).P("action", actionId), Title: title, Comment: comment, Icon: iconName, Relation: relation.Delete})
}*/
