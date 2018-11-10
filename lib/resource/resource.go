// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"
)

type GetHandler interface {
	GET(w http.ResponseWriter, r *http.Request)
}

type PostHandler interface {
	POST(w http.ResponseWriter, r *http.Request)
}

type PatchHandler interface {
	PATCH(w http.ResponseWriter, r *http.Request)
}

type DeleteHandler interface {
	DELETE(w http.ResponseWriter, r *http.Request)
}

type Resource interface {
	GetSelf() StandardizedPath
	GetMt() MediaType
}

type Executer func()




