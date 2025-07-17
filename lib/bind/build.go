package bind

import (
	"cmp"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/surlykke/RefudeServices/lib/entity"
	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/repo"
	"github.com/surlykke/RefudeServices/lib/response"
)

func ServeFunc(path string, function any, tags ...string) {
	http.HandleFunc(path, buildHandler(function, tags...))
}

func ServeMap[K cmp.Ordered, V entity.Servable](prefix string, m *repo.SyncMap[K, V]) {
	m.SetPrefix(prefix)
	ServeFunc("GET "+prefix+"{id...}", m.DoGetSingle, `path:"id"`)
	ServeFunc("GET "+prefix+"{$}", m.DoGetAll)
	ServeFunc("POST "+prefix+"{id...}", m.DoPost, `path:"id"`, `query:"action"`)
}

func buildHandler(function any, tags ...string) func(w http.ResponseWriter, r *http.Request) {
	var functionVal = reflect.ValueOf(function)
	var funcType = functionVal.Type()
	if funcType.Kind() != reflect.Func {
		panic("Not a function")
	}
	if funcType.NumOut() != 1 || funcType.Out(0) != reflect.TypeOf(response.Response{}) {
		panic("function does not return respond2.Response")
	}
	if funcType.NumIn() != len(tags) {
		panic("Number of tags does not match number of function parameters")
	}
	var deserializers = make([]deserializer, len(tags))
	for i, tag := range tags {

		var pb = readTag(reflect.StructTag(tag))
		if pb.bindingType == "body" {
			deserializers[i] = makeJsonBodyDeserializer(funcType.In(i))
		} else {
			var conv = getConverter(funcType.In(i))
			deserializers[i] = makeParmDeserializer("query" == pb.bindingType, pb.name, pb.required, pb.defaultValue, conv)
		}
	}

	return makeServer(functionVal, deserializers)
}

type deserializer func(r *http.Request) (reflect.Value, error)

func makeJsonBodyDeserializer(target reflect.Type) deserializer {
	return func(r *http.Request) (reflect.Value, error) {
		var valPtr = reflect.New(target)
		if bytes, err := io.ReadAll(r.Body); err != nil {
			return reflect.Value{}, err
		} else {
			err = json.Unmarshal(bytes, valPtr.Interface())
			return valPtr.Elem(), err
		}
	}
}

func makeParmDeserializer(fromQuery bool, name string, required bool, _default string, conv converter) deserializer {
	return func(r *http.Request) (reflect.Value, error) {
		var val string
		if fromQuery {
			if r.URL.Query().Has(name) {
				val = r.URL.Query()[name][0] // TODO lists
			}
		} else {
			val = r.PathValue(name)
		}
		if val == "" {
			val = _default
		}
		if val == "" && required {
			return reflect.Value{}, errors.New("query parameter '" + name + "' required and not given")
		}
		return conv(val)
	}

}

func makeServer(handlerFunction reflect.Value, deserializers []deserializer) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var values = make([]reflect.Value, len(deserializers), len(deserializers))
		var errs = make([]error, 0)
		for i := range values {
			if val, err := deserializers[i](r); err != nil {
				errs = append(errs, err)
			} else {
				values[i] = val
			}
		}
		if len(errs) > 0 {
			var joinedErrors = errors.Join(errs...)
			log.Warn("Responding unprocessable entity:", joinedErrors)
			response.UnprocessableEntity(joinedErrors).Send(w)
		} else {
			handlerFunction.Call(values)[0].Interface().(response.Response).Send(w)
		}
	}
}

type parameterBinding struct {
	bindingType  string
	name         string
	required     bool
	defaultValue string
}

func readTag(tag reflect.StructTag) parameterBinding {
	var (
		pb             parameterBinding
		bindingDetails string
		ok             bool
	)

	if bindingDetails, ok = tag.Lookup("query"); ok {
		pb.bindingType = "query"
	} else if bindingDetails, ok = tag.Lookup("path"); ok {
		pb.bindingType = "path"
	} else if bindingDetails, ok = tag.Lookup("body"); ok {
		pb.bindingType = "body"
	} else {
		panic("tag should start with 'query', 'path' or 'body'")
	}

	var elements = strings.Split(bindingDetails, ",")
	if len(elements) == 0 {
		panic("value for " + pb.bindingType + " empty")
	}
	pb.name = elements[0]
	for i := 1; i < len(elements); i++ {
		if elements[i] == "required" {
			pb.required = true
		} else if strings.HasPrefix(elements[i], "default=") {
			pb.defaultValue = elements[i][8:]
		} else {
			panic("only 'required' or 'default=<value>' allowed as qualifiers")
		}
	}
	if pb.bindingType == "body" {
		if pb.name != "json" {
			panic("Only 'json' supported for 'body:'")
		} else if pb.required || pb.defaultValue != "" {
			panic("'required' or 'default' not allowed for 'body:'")
		}
	} else {
		if pb.required && pb.defaultValue != "" {
			panic("'default' cannot be given when 'required'")
		}
	}
	return pb
}
