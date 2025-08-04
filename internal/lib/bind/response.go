// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package bind

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Status  int
	Headers http.Header
	Body    []byte
}

func (this Response) Send(w http.ResponseWriter) {
	var bodylen = fmt.Sprintf("%d", len(this.Body))
	w.Header().Add("Content-Length", bodylen)

	for headerName, headerValues := range this.Headers {
		for _, headerValue := range headerValues {
			w.Header().Add(headerName, headerValue)
		}
	}
	w.WriteHeader(this.Status)
	w.Write(this.Body)
}

func Ok() Response {
	return Response{Status: http.StatusOK}
}

func NotFound() Response {
	return Response{Status: http.StatusNotFound}
}

func NotAllowed() Response {
	return Response{Status: http.StatusMethodNotAllowed}
}

func UnprocessableEntity(err error) Response {
	return Response{Status: http.StatusUnprocessableEntity, Body: ToJson(err)}
}

func ServerError(err error) Response {
	return Response{Status: http.StatusInternalServerError, Body: []byte(err.Error())}
}

func Accepted() Response {
	return Response{Status: http.StatusAccepted}
}

func NotModified() Response {
	return Response{Status: http.StatusNotModified}
}

func PreconditionFailed() Response {
	return Response{Status: http.StatusPreconditionFailed}
}

func Json(data any) Response {
	return Response{Status: http.StatusOK, Headers: http.Header{"Content-Type": {"application/json"}}, Body: ToJson(data)}
}

func Html(html []byte) Response {
	return Response{Status: http.StatusOK, Headers: http.Header{"Content-Type": {"text/html"}}, Body: html}
}

func Image(contentType string, bytes []byte) Response {
	return Response{Status: http.StatusOK, Headers: http.Header{"Content-Type": {contentType}}, Body: bytes}
}

// We don't care about embedding in html, so no escaping
// (The standard encoder escapes '&', which is annoying when having links in json)
func ToJson(res any) []byte {
	var buf = bytes.NewBuffer([]byte{})
	var encoder = json.NewEncoder(buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(res); err != nil {
		panic(fmt.Sprintln(err))
	}
	return buf.Bytes()
}

func ToPrettyJson(res any) []byte {
	if b, err := json.MarshalIndent(res, "", "    "); err != nil {
		panic(err)
	} else {
		return b
	}
}
