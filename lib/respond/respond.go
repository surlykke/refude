package respond

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
)

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

func ServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func Accepted(w http.ResponseWriter) {
	w.WriteHeader(http.StatusAccepted)
}

type VersionedResource interface {
	GetEtag() string
}

type RefudeResource interface {
	http.Handler
	VersionedResource
}

type StandardFormat struct {
	Self         string
	OnPost       string `json:",omitempty"`
	OnPatch      string `json:",omitempty"`
	OnDelete     string `json:",omitempty"`
	OtherActions string `json:",omitempty"`
	Type         string
	Title        string
	Comment      string      `json:",omitempty"`
	IconName     string      `json:",omitempty"`
	Data         interface{} `json:",omitempty"`
}

func AsJson(w http.ResponseWriter, data interface{}) {
	var json = ToJson(data)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
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
