// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package bind

import (
	"fmt"
	"reflect"
	"strconv"
)

type converter func(paramVal string) (reflect.Value, error)

func getConverter(t reflect.Type) (converter, error) {
	switch t.Kind() {
	case reflect.Bool:
		return toBool, nil
	case reflect.Int:
		return toInt, nil
	case reflect.Int8:
		return toInt8, nil
	case reflect.Int16:
		return toInt16, nil
	case reflect.Int32:
		return toInt32, nil
	case reflect.Int64:
		return toInt64, nil
	case reflect.Uint:
		return toUint, nil
	case reflect.Uint8:
		return toUint8, nil
	case reflect.Uint16:
		return toUint16, nil
	case reflect.Uint32:
		return toUint32, nil
	case reflect.Uint64:
		return toUint64, nil
	case reflect.Float32:
		return toFloat32, nil
	case reflect.Float64:
		return toFloat64, nil
	case reflect.String:
		return toString, nil
	default:
		return nil, fmt.Errorf("Parameter type not supported: %s", t.Kind())
	}
}

func toBool(paramVal string) (reflect.Value, error) {
	return reflect.ValueOf("true" == paramVal), nil
}

func toInt(paramVal string) (reflect.Value, error) {
	var i64, err = strconv.ParseInt(paramVal, 10, strconv.IntSize)
	return reflect.ValueOf(int(i64)), err
}

func toInt8(paramVal string) (reflect.Value, error) {
	var i64, err = strconv.ParseInt(paramVal, 10, 8)
	return reflect.ValueOf(int8(i64)), err
}

func toInt16(paramVal string) (reflect.Value, error) {
	var i64, err = strconv.ParseInt(paramVal, 10, 16)
	return reflect.ValueOf(int16(i64)), err
}

func toInt32(paramVal string) (reflect.Value, error) {
	var i64, err = strconv.ParseInt(paramVal, 10, 32)
	return reflect.ValueOf(int32(i64)), err
}

func toInt64(paramVal string) (reflect.Value, error) {
	var i64, err = strconv.ParseInt(paramVal, 10, 64)
	return reflect.ValueOf(i64), err
}

func toUint(paramVal string) (reflect.Value, error) {
	var ui64, err = strconv.ParseUint(paramVal, 10, 64 /*fixme*/)
	return reflect.ValueOf(uint(ui64)), err
}

func toUint8(paramVal string) (reflect.Value, error) {
	var ui64, err = strconv.ParseUint(paramVal, 10, 8)
	return reflect.ValueOf(uint8(ui64)), err
}

func toUint16(paramVal string) (reflect.Value, error) {
	var ui64, err = strconv.ParseUint(paramVal, 10, 16)
	return reflect.ValueOf(uint16(ui64)), err
}

func toUint32(paramVal string) (reflect.Value, error) {
	var ui64, err = strconv.ParseUint(paramVal, 10, 32)
	return reflect.ValueOf(uint32(ui64)), err
}

func toUint64(paramVal string) (reflect.Value, error) {
	var ui64, err = strconv.ParseUint(paramVal, 10, 64)
	return reflect.ValueOf(ui64), err
}

func toFloat32(paramVal string) (reflect.Value, error) {
	var f64, err = strconv.ParseFloat(paramVal, 32)
	return reflect.ValueOf(float32(f64)), err
}

func toFloat64(paramVal string) (reflect.Value, error) {
	var f64, err = strconv.ParseFloat(paramVal, 64)
	return reflect.ValueOf(f64), err
}

func toString(paramVal string) (reflect.Value, error) {
	return reflect.ValueOf(paramVal), nil
}
