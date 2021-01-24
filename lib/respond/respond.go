package respond

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
)

func Ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

func NotFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func NotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func UnprocessableEntity(w http.ResponseWriter, err error) {
	if body, err2 := json.Marshal(err.Error()); err2 == nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write(body)
	} else {
		panic(fmt.Sprintf("Cannot json-marshall %s", err.Error()))
	}
}

func ServerError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(err.Error()))
}

func Accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

func AcceptedAndThen(w http.ResponseWriter, f func()) {
	w.WriteHeader(http.StatusAccepted)
	f()
}

// -----

func AsJson(w http.ResponseWriter, data interface{}) {
	var json = ToJson(data)
	w.Header().Set("Content-Type", "application/vnd.refude+json")
	w.Write(json)
}

func AsPng(w http.ResponseWriter, pngData []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(pngData)
}

func ToJson(res interface{}) []byte {
	var bytes, err = json.Marshal(res)
	if err != nil {
		panic(fmt.Sprintln(err))
	}
	return bytes
}

func ToJsonAndEtag(res interface{}) ([]byte, string) {
	var bytes = ToJson(res)
	return bytes, fmt.Sprintf("\"%x\"", sha1.Sum(bytes))
}
