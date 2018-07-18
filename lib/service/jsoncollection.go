package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

const LinksMediaType mediatype.MediaType = "application/vnd.org.refude.links+json"

type JsonCollection interface {
	GetResource(path StandardizedPath) *resource.JsonResource
	GetAll() []*resource.JsonResource
}

type Links map[mediatype.MediaType][]StandardizedPath

func (Links) GetSelf() string {
	return "/links"
}

func (Links) GetMt() mediatype.MediaType {
	return LinksMediaType
}



