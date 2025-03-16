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
	ServeFunc("GET "+prefix+"{id...}", m.DoGetSingle, "type=path,name=id")
	ServeFunc("GET "+prefix+"{$}", m.DoGetAll)
	ServeFunc("POST "+prefix+"{id...}", m.DoPost, "type=path,name=id", "name=action")
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
	var deserializers = make([]deserializer, len(tags), len(tags))
	for i, tag := range tags {
		var tagData = readTag(tag)
		var inputType = tagData["type"]
		if inputType == "json" {
			deserializers[i] = makeJsonBodyDeserializer(funcType.In(i))
		} else {
			var fromQuery = inputType == "query"
			var name = tagData["name"]
			var required = "true" == tagData["required"]
			var _default = tagData["default"]
			var conv = getConverter(funcType.In(i))
			deserializers[i] = makeParmDeserializer(fromQuery, name, required, _default, conv)
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
		for i := 0; i < len(values); i++ {
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

func readTag(tag string) map[string]string {
	var attrs = map[string]string{
		"type":     "query",
		"required": "false",
		"default":  "",
	}
	for _, pair := range strings.Split(tag, ",") {
		var keyVal = strings.Split(pair, "=")
		if len(keyVal) != 2 {
			panic("tag should be a comma-separated list of key-value pairs separated by '=' <key>=<val> : " + tag)
		}
		attrs[keyVal[0]] = keyVal[1]
	}

	for key, val := range attrs {
		switch key {
		case "type":
			switch val {
			case "query", "path":
				if attrs["name"] == "" {
					panic("'name' must be given for  'query' and 'path'")
				}
			case "json":
				if attrs["name"] != "" || attrs["required"] != "false" || attrs["default"] != "" {
					panic("attributes 'name', 'required' and 'default' should not be given for type 'json'")
				}
			default:
				panic("only types 'query', 'path' and 'json' supported")
			}
		case "name": // handled above
		case "required":
			if val != "true" && val != "false" {
				panic("required must be 'true' or 'false'")
			}
		case "default": // Anything goes
		default:
			panic("Unknown attibute: '" + key + "'")
		}
	}
	return attrs
}
