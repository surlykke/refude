package resserv

import (
	"github.com/surlykke/RefudeServices/lib/repo"
)

var sinks = make(map[string]chan repo.ResourceRequest)


/*func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var path = r.URL.Path
	var sink = getSink(path)
	if sink == nil {
		respond.NotFound(w)
	} else {
		var back = make(chan resource.RankedResource)
		if strings.HasSuffix(path, "/") {
		}
	}
}

func getSink(path string) chan repo.ResourceRequest {
	return nil // FIXE
}*/
