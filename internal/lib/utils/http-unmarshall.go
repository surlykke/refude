package utils

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

func QueryParam(r *http.Request, paramName string) string {
	var values = r.URL.Query()[paramName]
	if len(values) > 0 {
		return values[0]
	} else {
		return ""
	}
}

func Convert[T any](strVal string, t *T) error {
	var val = reflect.ValueOf(t)
	var kind = val.Elem().Kind()
	fmt.Println("Attempt conversion of", strVal, "to", kind)
	switch kind {
	case reflect.Bool:
		val.SetBool("true" == strVal)
	case reflect.String:
		val.Elem().SetString(strVal)
	case reflect.Int:
		if v, err := strconv.ParseInt(strVal, 10, strconv.IntSize); err != nil {
			return err
		} else {
			val.Elem().SetInt(v)
		}
	case reflect.Int8:
		if v, err := strconv.ParseInt(strVal, 10, 8); err != nil {
			return err
		} else {
			val.Elem().SetInt(v)
		}
	case reflect.Int16:
		if v, err := strconv.ParseInt(strVal, 10, 16); err != nil {
			return err
		} else {
			val.Elem().SetInt(v)
		}
	case reflect.Int32:
		if v, err := strconv.ParseInt(strVal, 10, 32); err != nil {
			return err
		} else {
			val.Elem().SetInt(v)
		}
	case reflect.Int64:
		if v, err := strconv.ParseInt(strVal, 10, 64); err != nil {
			return err
		} else {
			val.Elem().SetInt(v)
		}
	case reflect.Uint:
		if v, err := strconv.ParseUint(strVal, 10, strconv.IntSize); err != nil {
			return err
		} else {
			val.Elem().SetUint(v)
		}
	case reflect.Uint8:
		if v, err := strconv.ParseUint(strVal, 10, 8); err != nil {
			return err
		} else {
			val.Elem().SetUint(v)
		}
	case reflect.Uint16:
		if v, err := strconv.ParseUint(strVal, 10, 16); err != nil {
			return err
		} else {
			val.Elem().SetUint(v)
		}
	case reflect.Uint32:
		if v, err := strconv.ParseUint(strVal, 10, 32); err != nil {
			return err
		} else {
			val.Elem().SetUint(v)
		}
	case reflect.Uint64:
		if v, err := strconv.ParseUint(strVal, 10, 64); err != nil {
			return err
		} else {
			val.Elem().SetUint(v)
		}
	case reflect.Float32:
		if v, err := strconv.ParseFloat(strVal, 32); err != nil {
			return err
		} else {
			val.Elem().SetFloat(v)
		}
	case reflect.Float64:
		if v, err := strconv.ParseFloat(strVal, 64); err != nil {
			return err
		} else {
			val.Elem().SetFloat(v)
		}

	default:
		panic(fmt.Sprintf("unexpected reflect.Kind: %#v", kind))
	}
	return nil
}
