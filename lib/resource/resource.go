// Copyright (c) 2017 Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package resource

import (
	"net/http"
	"encoding/json"
	"crypto/sha1"
	"errors"
)

type MediaType string

type Resource interface {
	GET(w http.ResponseWriter, r *http.Request)
	PATCH(w http.ResponseWriter, r *http.Request)
	POST(w http.ResponseWriter, r *http.Request)
	DELETE(w http.ResponseWriter, r *http.Request)
	MediaType() MediaType
	ETag() string
	Update() Resource
}

type DefaultResource struct{}

func (d *DefaultResource) GET(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *DefaultResource) PATCH(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *DefaultResource) POST(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *DefaultResource) DELETE(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func (d *DefaultResource) MediaType() MediaType {
	panic("Method MediaType() not defined")
}

func (d *DefaultResource) ETag() string {
	return ""
}

func (d *DefaultResource) Update() Resource {
	return nil
}

type ByteResource struct {
	DefaultResource
	mediaType MediaType
	etag      string
	bytes     []byte
}

func MakeByteResource(mediaType MediaType) ByteResource {
	return ByteResource{mediaType: mediaType}
}

func (j *ByteResource) SetBytes(bytes []byte) {
	j.bytes = bytes
	hash := sha1.New()
	hash.Write(bytes)
	j.etag = string(hash.Sum(nil))
}

func (j *ByteResource) GetBytes() []byte {
	return j.bytes
}


func (j *ByteResource) GET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", string(j.mediaType))
	w.Write(j.bytes)
}

func (j *ByteResource) MediaType() MediaType {
	if j.mediaType == "" {
		panic("No mediatype")
	}

	return j.mediaType
}


func ToJSon(res interface{}) []byte {
	if bytes, err := json.Marshal(res); err != nil {
		panic("Could not json-marshal")
	} else {
		return bytes
	}
}

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}
/**
 * Errors if r.URL.Query contains parameters not in params
 * Errors if any of paramNames has more than one value
 * Returns value from query if there, "" if not
 */
func GetSingleParams(r *http.Request, paramNames...string) (map[string]string, error) {
	for queryParam, _ := range r.URL.Query() {
		var ok = false
		for _,paramName := range paramNames {
			if queryParam == paramName {
				ok = true
				break
			}
		}

		if !ok {
			return nil, errors.New("Unexpected query parameter:" + queryParam)
		}
	}
	var result = make(map[string]string)

	for _,paramName := range paramNames {
		if len(r.URL.Query()[paramName]) > 1 {
			return nil, errors.New("Multible values for " + paramName)
		} else if len(r.URL.Query()[paramName]) == 1 {
			result[paramName] = r.URL.Query()[paramName][0]
		}

	}

	return result, nil
}
