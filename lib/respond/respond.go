package respond

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
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
	NoDisplay    bool        `json:"-"`
	Rank         int         `json:"-"`
}

type StandardFormatList []*StandardFormat

type rankSortable StandardFormatList

func (rs rankSortable) Len() int { return len(rs) }

func (rs rankSortable) Less(i int, j int) bool {
	if rs[i].Rank == rs[j].Rank {
		return rs[i].Self < rs[j].Self
	} else {
		return rs[i].Rank < rs[j].Rank
	}
}

func (rs rankSortable) Swap(i int, j int) { rs[i], rs[j] = rs[j], rs[i] }

type pathSortable StandardFormatList

func (ps pathSortable) Len() int               { return len(ps) }
func (ps pathSortable) Less(i int, j int) bool { return ps[i].Rank < ps[j].Rank }
func (ps pathSortable) Swap(i int, j int)      { ps[i], ps[j] = ps[j], ps[i] }

func (sf *StandardFormat) Ranked(rank int) *StandardFormat {
	sf.Rank = rank
	return sf
}

func (sf StandardFormatList) SortByRank() StandardFormatList {
	sort.Sort(rankSortable(sf))
	return sf
}

func (sf StandardFormatList) SortByPath() StandardFormatList {
	sort.Sort(pathSortable(sf))
	return sf
}

// -----

func AsJson(w http.ResponseWriter, r *http.Request, data interface{}) {
	if r.Method == "GET" {
		var json = ToJson(data)
		w.Header().Set("Content-Type", "application/json")
		w.Write(json)
	} else {
		NotAllowed(w)
	}
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
