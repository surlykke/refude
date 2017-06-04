package common

import (
	"encoding/json"
	"net/http"
)

func ServeAsJson(w http.ResponseWriter, r *http.Request, i interface{}) {
	bytes, err := json.Marshal(i)
	if err != nil {
		panic("Could not json-marshal")
	};
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

