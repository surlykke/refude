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


func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}
