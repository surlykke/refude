// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package respond

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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

func NotModified(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotModified)
}

func PreconditionFailed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusPreconditionFailed)
}

func AsJson(w http.ResponseWriter, data interface{}) {
	var json = ToJson(data)
	w.Header().Set("Content-Type", "application/vnd.refude+json")
	w.Write(json)
}

func AsHtml(w http.ResponseWriter, html []byte) {
	w.Header().Set("Content-Type", "text/html")
	w.Write(html)
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

// We don't care about embedding in html, so no escaping
// (The standard encoder escapes '&', which is annoying when having links in json)
func ToJson(res interface{}) []byte {
	var buf = bytes.NewBuffer([]byte{})
	var encoder = json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(res); err != nil {
		panic(fmt.Sprintln(err))
	}
	return buf.Bytes()
}
