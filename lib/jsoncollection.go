package lib

const LinksMediaType MediaType = "application/vnd.org.refude.links+json"

type JsonCollection interface {
	GetResource(path StandardizedPath) *JsonResource
	GetAll() []*JsonResource
}

type Links map[MediaType][]StandardizedPath

func (Links) GetSelf() StandardizedPath {
	return "/links"
}

func (Links) GetMt() MediaType {
	return LinksMediaType
}



