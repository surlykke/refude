package handle

import "net/http"

type Get interface {
	GET(w http.ResponseWriter, r *http.Request)
}

type Post interface {
	POST(w http.ResponseWriter, r *http.Request)
}

type Delete interface {
	DELETE(w http.ResponseWriter, r *http.Request)
}

type Patch interface {
	PATCH(w http.ResponseWriter, r *http.Request)
}
