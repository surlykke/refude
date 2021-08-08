package respond

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/surlykke/RefudeServices/lib/resource"
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

func AsJson(w http.ResponseWriter, data interface{}) {
	var json = ToJson(data)
	w.Header().Set("Content-Type", "application/vnd.refude+json")
	w.Write(json)
}

func writeOrPanic(w io.Writer, byteArrArr ...[]byte) {
	for _, byteArr := range byteArrArr {
		for len(byteArr) > 0 {
			if i, err := w.Write(byteArr); err != nil {
				panic(err)
			} else {
				byteArr = byteArr[i:]
			}
		}
	}
}

// TODO doc
func ResourceAsJson(w http.ResponseWriter, links []resource.Link, refudeType string, res interface{}) {
	w.Header().Set("Content-Type", "application/vnd.refude+json")
	var linksJson = bytes.TrimSpace(ToJson(links))
	var resJson = bytes.TrimSpace(ToJson(res))
	if resJson[0] != '{' || len(resJson) < 2 {
		panic("res must serialize to a Json object")
	}
	writeOrPanic(w, []byte(`{"_links":`), linksJson, []byte(`,"refudeType":"`), []byte(refudeType), []byte{'"'})
	if len(resJson) > 2 {
		writeOrPanic(w, []byte{','})
	}
	writeOrPanic(w, resJson[1:len(resJson)-1], []byte{'}'})
}

func AsPng(w http.ResponseWriter, pngData []byte) {
	w.Header().Set("Content-Type", "image/png")
	w.Write(pngData)
}

// We don't care about embedding in html, so no escaping
// (The standard encoder escapes '&' which was a nuisance)
func ToJson(res interface{}) []byte {
	var buf = bytes.NewBuffer([]byte{})
	var encoder = json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(res); err != nil {
		panic(fmt.Sprintln(err))
	}
	return buf.Bytes()
}
