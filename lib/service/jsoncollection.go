package service

import (
	"github.com/surlykke/RefudeServices/lib/resource"
	"github.com/surlykke/RefudeServices/lib/mediatype"
)

type Links map[mediatype.MediaType][]string

type JsonCollection interface {
	GetResource(path StandardizedPath) *resource.JsonResource
	GetAll() []*resource.JsonResource
	GetLinks() Links
}