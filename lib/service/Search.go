package service

import (
	"net/url"
	"strings"
	"reflect"
)

type matcher func(res interface{}) bool

type NormalizedParameter struct{
	splittedKeys []string
	upcasedValues []string
}

type NormalizedQuery map[string]NormalizedParameter

func Search(query url.Values) []interface{} {
	normalizedQuery := make(NormalizedQuery)
	for parameterName, parameterValues := range query {
		upcasedValues :=  make([]string, 0, len(parameterValues))
		for _, parameterValue := range parameterValues {
			upcasedValues = append(upcasedValues, strings.ToUpper(parameterValue))
		}

		normalizedQuery[parameterName] = NormalizedParameter{strings.Split(parameterName, ","), upcasedValues}
	}

	var result = make([]interface{}, 0, 100)
	for _, res := range resources {
		if matchAllConditions(res, normalizedQuery)  {
			result = append(result, res)
		}
	}

	return result
}

func matchAllConditions(res interface{}, normalizedQuery NormalizedQuery) bool {
	for _, np := range normalizedQuery {
		if ! matchOneCombination(res, np.splittedKeys, np.upcasedValues) {
			return false
		}
	}

	return true
}

func matchOneCombination(res interface{}, keys []string, upcaseValues []string) bool {
	for _,key := range keys {
		fieldValue, ok := extractUpcaseFieldValue(res, key)
		if ok {
			for _, value := range upcaseValues {
				var not = value[0:1] == "!"
				if not {
					value = value[1:]
				}

				if value[0:1] == "~" && (not != strings.Contains(fieldValue, value[1:])) {
					return true
				} else if value == fieldValue {
					return true
				}
			}
		}
	}

	return false
}

func extractUpcaseFieldValue(res interface{}, fieldName string) (string, bool) {
	v := reflect.ValueOf(res)

	for v.Kind() == reflect.Ptr {
		v = reflect.Indirect(v)
	}

	if v.Kind() == reflect.Struct {
		f := v.FieldByName(fieldName)
		if f.Kind() == reflect.String {
			return strings.ToUpper(f.String()), true
		}
	}

	return "", false
}

