package bind

import (
	"errors"
	"net/http"
	"reflect"

	"github.com/surlykke/RefudeServices/lib/log"
	"github.com/surlykke/RefudeServices/lib/response"
)

func buildHandler(function any, bindingspec []binding) func(w http.ResponseWriter, r *http.Request) {
	var functionVal = reflect.ValueOf(function)
	var funcType = functionVal.Type()
	if funcType.Kind() != reflect.Func {
		panic("Not a function")
	}
	if funcType.NumOut() != 1 || funcType.Out(0) != reflect.TypeOf(response.Response{}) {
		panic("function does not return respond2.Response")
	}
	if funcType.NumIn() != len(bindingspec) {
		panic("Length of bindingspec does not match number of function parameters")
	}
	var readers = make([]reader, len(bindingspec), len(bindingspec))
	var converters = make([]converter, len(bindingspec), len(bindingspec))
	for i, bs := range bindingspec {
		if bs.query != "" {
			readers[i] = makeQueryReader(bs.query, bs.req, bs.def)
		} else if bs.path != "" {
			readers[i] = makePathReader(bs.path)
		} else {
			readers[i] = bodyReader
		}
		if bs.body == "json" {
			converters[i] = makeJsonConverter(funcType.In(i))
		} else {
			converters[i] = getConverter(funcType.In(i))
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var values = make([]reflect.Value, len(readers), len(readers))
		var errs = make([]error, 0)
		for i := 0; i < len(values); i++ {
			if strVal, err := readers[i](r); err != nil {
				errs = append(errs, err)
			} else if val, err := converters[i](strVal); err != nil {
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
			functionVal.Call(values)[0].Interface().(response.Response).Send(w)
		}
	}
}
