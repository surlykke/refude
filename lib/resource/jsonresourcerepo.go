package resource

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/requests"
)

type ResourceRepo interface {
	Get(path string) interface{}
}

type JsonResourceServer struct {
	ResourceRepo
}

func MakeJsonResourceServer(repo ResourceRepo) JsonResourceServer {
	return JsonResourceServer{ResourceRepo: repo}
}

func (jrs JsonResourceServer) fetch(r *http.Request) (interface{}, error) {
	var res = jrs.Get(r.URL.EscapedPath())
	if res == nil {
		return nil, nil
	}

	if list, ok := res.([]interface{}); ok {
		var matcher, err = requests.GetMatcher(r)
		if err != nil {
			return nil, err
		}

		if matcher != nil {
			var filtered = 0
			for _, res := range list {
				if matcher(res) {
					list[filtered] = res
					filtered++
				}
			}
			list = list[0:filtered]
		}

		

		var _, brief = r.URL.Query()["brief"]
		if brief {
			var pathList = make([]string, len(list))
			for i := 0; i < len(pathList); i++ {
				pathList[i] = list[i].(Selfie).GetSelf()
			}
			return pathList, nil
		} else {
			return list, nil
		}

	} else {
		return res, nil
	}
}

func (jrs JsonResourceServer) ServeHTTP(w http.ResponseWriter, r *http.Request) bool {
	res, err := jrs.fetch(r)

	if err != nil {
		requests.ReportUnprocessableEntity(w, err)
		return true
	}

	if res == nil {
		return false
	}

	if handler, ok := res.(http.Handler); ok {
		handler.ServeHTTP(w, r)
		return true
	}

	if r.Method == "GET" {
		var bytes, etag = ToBytesAndEtag(res)
		if statusCode := requests.CheckEtag(r, etag); statusCode != 0 {
			w.WriteHeader(statusCode)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("ETag", etag)
			_, _ = w.Write(bytes)
		}
	} else if posthandler, ok := res.(PostHandler); ok && r.Method == "POST" {
		posthandler.POST(w, r)
	} else if patchHandler, ok := res.(PatchHandler); ok && r.Method == "PATCH" {
		patchHandler.PATCH(w, r)
	} else if deleteHandler, ok := res.(DeleteHandler); ok && r.Method == "DELETE" {
		deleteHandler.DELETE(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}

	return true
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
