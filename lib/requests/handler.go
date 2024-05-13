package requests

import (
	"net/http"
)

type Request struct {
	W    http.ResponseWriter
	R    *http.Request
	Done chan struct{}
}

type Handler chan Request

func (this Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var done = make(chan struct{})
	this <- Request{w, r, done}
	<-done
}

// To be called from receiving goroutine
func Handle(req Request, httpHandler func(w http.ResponseWriter, r *http.Request)) {
	httpHandler(req.W, req.R)
	req.Done <- struct{}{}
}
