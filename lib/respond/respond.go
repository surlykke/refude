package respond

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"github.com/surlykke/RefudeServices/lib/requests"
	"github.com/surlykke/RefudeServices/lib/searchutils"
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
	NoDisplay    bool        `json:"-"`
	Rank         int         `json:"-"`
}

type StandardFormatList []*StandardFormat

func (sfl StandardFormatList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		AsJson2(w, sfl.Filter(requests.Term(r)).Sort())
	} else {
		NotAllowed(w)
	}
}

func (sfl StandardFormatList) Rank(rank int) {
	for _, sf := range sfl {
		sf.Rank = rank
	}
}

func (sfl StandardFormatList) Filter(term string) StandardFormatList {
	var filtered = make(StandardFormatList, 0, len(sfl))
	for _, sf := range sfl {
		if sf.NoDisplay {
			continue
		}
		var rank = searchutils.SimpleRank(sf.Title, sf.Comment, term)
		if rank <= -1 {
			continue
		}

		sf.Rank = sf.Rank + rank
		filtered = append(filtered, sf)
	}

	return filtered
}

func (sfl StandardFormatList) Len() int { return len(sfl) }

func (sfl StandardFormatList) Less(i int, j int) bool {
	if sfl[i].Rank == sfl[j].Rank {
		return sfl[i].Self < sfl[j].Self
	} else {
		return sfl[i].Rank < sfl[j].Rank
	}
}

func (sfl StandardFormatList) Swap(i int, j int) { sfl[i], sfl[j] = sfl[j], sfl[i] }

func (sf StandardFormatList) Sort() StandardFormatList {
	sort.Sort(sf)
	return sf
}

// -----

func AsJson2(w http.ResponseWriter, data interface{}) {
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
