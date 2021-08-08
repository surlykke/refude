package resource

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/link"
)

/**
 * A resource is something that is 'linkable' and has a type.
 */
type Resource interface {
	// The resources links. Should be not empty and the first link should be the self-link
	Links() link.List
	RefudeType() string
}

type Postable interface {
	DoPost(w http.ResponseWriter, r *http.Request)
}

type Deleteable interface {
	DoDelete(w http.ResponseWriter, r *http.Request)
}
