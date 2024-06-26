package repo

import (
	"github.com/surlykke/RefudeServices/lib/resource"
)

func FindSingle(path string) resource.Resource {
	var resources = find(ByPath, path)
	if len(resources) == 1 {
		return resources[0]
	} else if len(resources) == 0 {
		return nil
	} else {
		panic("More than one resource under " + path)
	}
}

func FindList(prefix string) []resource.Resource {
	return find(ByPathPrefix, prefix)
}

func DoSearch(term string) []resource.Resource {
	return find(Search, term)
}

func find(reqType ResourceRequestType, data string) []resource.Resource {
	var req = ResourceRequest{
		ReqType: reqType,
		Data:    data,
		Replies: make(chan resource.RankedResource),
	}
	requests <- req

	var ranked = make(resource.RRList, 0, 100)
	for rres := range req.Replies {
		ranked = append(ranked, rres)
	}
	return ranked.GetResources()
}
