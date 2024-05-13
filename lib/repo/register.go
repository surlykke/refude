package repo

import (
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

type ResourceRequestType uint8

const (
	ByPathPrefix ResourceRequestType = iota
	ByPath
	Search
)

type ResourceRequest struct {
	ReqType ResourceRequestType
	Data    string
	Replies chan resource.RankedResource
	Wg      *sync.WaitGroup
}

type RequestChan chan ResourceRequest

func MakeAndRegisterRequestChan() RequestChan {
	var ch = make(RequestChan)
	register = append(register, ch)
	return ch
}

var register = make([]RequestChan, 0, 10)

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
		Wg:      &sync.WaitGroup{},
	}
	for _, reqChan := range register {
		req.Wg.Add(1)
		reqChan <- req
	}

	var done = make(chan struct{})
	go func() {
		req.Wg.Wait()
		close(done)
	}()

	var ranked = make(resource.RRList, 0, 100)
forloop:
	for {
		select {
		case res := <-req.Replies:
			ranked = append(ranked, res)
		case <-done:
			break forloop
		}
	}
	return ranked.GetResourcesSorted()
}
