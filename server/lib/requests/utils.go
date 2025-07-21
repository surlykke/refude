// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package requests

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
)

func GetSingleQueryParameter(r *http.Request, parameterName string, fallbackValue string) string {
	if len(r.URL.Query()[parameterName]) == 0 {
		return fallbackValue
	} else {
		return r.URL.Query()[parameterName][0]
	}
}

func GetPosInt(r *http.Request, parameterName string) (uint, error) {
	var paramValue = GetSingleQueryParameter(r, parameterName, "")
	if intVal, err := strconv.Atoi(paramValue); err != nil {
		return 0, err
	} else if intVal <= 0 {
		return 0, errors.New(fmt.Sprintf("'%s' should be positive", parameterName))
	} else {
		return uint(intVal), nil
	}
}
