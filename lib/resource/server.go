package resource

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type ResourceRepo interface {
	Get(path string) Res
	LongGet(path string, etagList string) Res
}

type Server struct {
	ResourceRepo
}

func MakeServer(repo ResourceRepo) Server {
	return Server{ResourceRepo: repo}
}

func (jrs Server) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	var path = r.URL.Path
	var res Res
	if r.Method == "GET" && requests.HaveParam(r, "longpoll") {
		res = jrs.LongGet(path, r.Header.Get("If-None-Match"))
	} else {
		res = jrs.Get(path)
	}

	if res == nil {
		return false
	} else {
		res.ServeHTTP(w, r)
		return true
	}
}

func ToJson(res interface{}) []byte {
	var bytes, err = json.Marshal(res)
	if err != nil {
		panic(fmt.Sprintln(err))
	}

	return bytes
}

func ToBytesAndEtag(res interface{}) ([]byte, string) {
	var bytes []byte
	var etag string
	var err error
	if bytes, err = json.Marshal(res); err != nil {
		panic(fmt.Sprintln(err))
	}

	return bytes, etag
}
