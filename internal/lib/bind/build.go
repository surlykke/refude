// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package bind

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

func MakeAdapter(function any, bindings ...binding) (func(r *http.Request) Response, error) {
	var functionVal = reflect.ValueOf(function)
	var funcType = functionVal.Type()

	if funcType.Kind() != reflect.Func {
		return nil, errors.New("Not a function")
	}

	if funcType.NumOut() != 1 || funcType.Out(0) != reflect.TypeOf(Response{}) {
		return nil, errors.New("function does not have a single return value or type Response")
	}

	if funcType.NumIn() != len(bindings) {
		return nil, errors.New("Number of bindings does not match number of function parameters")
	}

	var deserializers = make([]deserializer, len(bindings))
	var errs = []error{}

	for i, b := range bindings {
		if d, err := makeDeserializer(b, funcType.In(i)); err != nil {
			errs = append(errs, err)
		} else {
			deserializers[i] = d
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	} else {
		return func(r *http.Request) Response {
			var values = make([]reflect.Value, len(deserializers))
			var errs = []error{}
			for i := range values {
				if val, err := deserializers[i](r); err != nil {
					errs = append(errs, err)
				} else {
					values[i] = val
				}
			}
			if len(errs) > 0 {
				fmt.Println("Return unprocessable entity", errs)
				return UnprocessableEntity(errors.Join(errs...))
			} else {
				return functionVal.Call(values)[0].Interface().(Response)
			}
		}, nil
	}
}

func Adapter(function any, bindings ...binding) func(r *http.Request) Response {
	if adapter, err := MakeAdapter(function, bindings...); err != nil {
		panic(err)
	} else {
		return adapter
	}
}

func ServeFunc(function any, bindings ...binding) func(w http.ResponseWriter, r *http.Request) {
	var adapter = Adapter(function, bindings...)
	return func(w http.ResponseWriter, r *http.Request) {
		adapter(r).Send(w)
	}
}

type deserializer func(r *http.Request) (reflect.Value, error)

func makeDeserializer(b binding, _type reflect.Type) (deserializer, error) {
	if b.kind == body {
		if b.qualifier == "json" {
			return func(r *http.Request) (reflect.Value, error) {
				var valPtr = reflect.New(_type)
				if bytes, err := io.ReadAll(r.Body); err != nil {
					return reflect.Value{}, err
				} else {
					err = json.Unmarshal(bytes, valPtr.Interface())
					return valPtr.Elem(), err
				}
			}, nil
		} else {
			return nil, errors.New("Unrecognized body type: " + b.qualifier + ". Only 'json' is supported")
		}
	} else {
		var conv, err = getConverter(_type)
		if err != nil {
			return nil, err
		}
		if b.kind == query {
			return func(r *http.Request) (reflect.Value, error) {
				var val string
				if r.URL.Query().Has(b.qualifier) {
					val = r.URL.Query()[b.qualifier][0] // TODO: lists
				} else if !b.optional {
					return reflect.Value{}, errors.New("query parameter '" + b.qualifier + "' required and not given")
				} else {
					val = b.defaultValue
				}
				return conv(val)
			}, nil
		} else if b.kind == path {
			return func(r *http.Request) (reflect.Value, error) {
				return conv(r.PathValue(b.qualifier))
			}, nil
		} else {
			panic("Should not happen")
		}
	}
}

type parameterBinding struct {
	bindingType  string
	name         string
	required     bool
	defaultValue string
}
