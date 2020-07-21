package searchutils

import (
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

func Term(r *http.Request) string {
	return requests.GetSingleQueryParameter(r, "term", "")
}
