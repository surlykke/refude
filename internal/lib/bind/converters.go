// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package bind

import (
	"log"
	"reflect"
	"strconv"
)

type converter func(paramVal string) (reflect.Value, error)

func getConverter(t reflect.Type) converter {
	switch t.Kind() {
	case reflect.Bool:
		return toBool
	case reflect.Int:
		return toInt
	case reflect.Int8:
		return toInt8
	case reflect.Int16:
		return toInt16
	case reflect.Int32:
		return toInt32
	case reflect.Int64:
		return toInt64
	case reflect.Uint:
		return toUint
	case reflect.Uint8:
		return toUint8
	case reflect.Uint16:
		return toUint16
	case reflect.Uint32:
		return toUint32
	case reflect.Uint64:
		return toUint64
	case reflect.Float32:
		return toFloat32
	case reflect.Float64:
		return toFloat64
	case reflect.String:
		return toString
	default:
		log.Panic("Parameter type not supported", t, t.Kind())
		return nil
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
	var f64, err = floatHelper(paramVal, 32)
	return reflect.ValueOf(float32(f64)), err
}

func toFloat64(paramVal string) (reflect.Value, error) {
	var f64, err = floatHelper(paramVal, 64)
	return reflect.ValueOf(f64), err
}

func toString(paramVal string) (reflect.Value, error) {
	return reflect.ValueOf(paramVal), nil
}

func intHelper(str string, bitsize int) (int64, error) {
	if str == "" {
		return 0, nil
	} else {
		return strconv.ParseInt(str, 10, bitsize)
	}
}

func uIntHelper(str string, bitsize int) (uint64, error) {
	if str == "" {
		return 0, nil
	} else {
		return strconv.ParseUint(str, 10, bitsize)
	}
}

func floatHelper(str string, bitsize int) (float64, error) {
	if str == "" {
		return 0, nil
	} else {
		return strconv.ParseFloat(str, bitsize)
	}
}
