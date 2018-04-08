package service

import (
	"net/url"
	"strings"
	"reflect"
	"fmt"
	"errors"
)

type matcher func(res interface{}) bool

type NormalizedParameter struct{
	splittedKeys []string
	upcasedValues []string
}

type NormalizedQuery map[string]NormalizedParameter

func Search(query url.Values) ([]interface{}, error) {
	fmt.Println("Search, q:", query["q"])
	if len(query["q"]) == 0 {
		return []interface{}{}, errors.New("No query given")
	}

	if m, err:= parseQuery(query["q"][0]); err == nil {
		var result = make([]interface{}, 0, 100)
		for _, res := range resources {
			if m(res) {
				result = append(result, res)
			}
		}

		return result, nil
	} else {
		return []interface{}{}, err
	}


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
				if value[0:1] == "~" && strings.Contains(fieldValue, value[1:]) {
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

