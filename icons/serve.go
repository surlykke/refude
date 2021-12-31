package icons

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/respond"
)

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		respond.NotAllowed(w)
	} else if name := requests.GetSingleQueryParameter(r, "name", ""); name == "" {
		respond.UnprocessableEntity(w, fmt.Errorf("Query parameter 'name' must be given, and not empty"))
	} else if strings.HasPrefix(name, "/") {
		http.ServeFile(w, r, name)
	} else if size, err := extractSize(r); err != nil {
		respond.UnprocessableEntity(w, err)
	} else if iconFilePath := findIconPath(name, size); iconFilePath == "" {
		respond.NotFound(w)
	} else {
		http.ServeFile(w, r, iconFilePath)
	}
}

func extractSize(r *http.Request) (uint32, error) {
	var size uint32 = 32

	if len(r.URL.Query()["size"]) > 0 {
		if size64, err := strconv.ParseUint(r.URL.Query()["size"][0], 10, 32); err != nil {
			return 0, errors.New("Invalid size given:" + r.URL.Query()["size"][0])
		} else {
			size = uint32(size64)
		}
	}

	return size, nil
}
