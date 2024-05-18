package repo

import (
	"sync"

	"github.com/surlykke/RefudeServices/lib/resource"
)

var register = make([]chan ResourceRequest, 0, 10)

var registrations = make(chan chan ResourceRequest)
var requests = make(chan ResourceRequest)

func Run() {
	for {
		select {
		case registration := <-registrations:
			register = append(register, registration)
		case sreq := <-requests:
			var req = ResourceRequest{
				ReqType: sreq.ReqType,
				Data:    sreq.Data,
				Replies: make(chan resource.RankedResource),
				Wg: &sync.WaitGroup{},
			}
			for _, source := range register {
				req.Wg.Add(1)
				source <- req
			}
			var done = make(chan struct{})
			go func() {
				req.Wg.Wait()
				close(done)
			}()

		forloop:
			for {
				select {
				case res := <-req.Replies:
					sreq.Replies <- res
				case <-done:
					break forloop
				}
			}
			close(sreq.Replies)
		}
	}
}

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

func MakeAndRegisterRequestChan() chan ResourceRequest {
	var ch = make(chan ResourceRequest)
	registrations <- ch
	return ch
}
