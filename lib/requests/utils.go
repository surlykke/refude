// Copyright (c) Christian Surlykke
//
// This file is part of the RefudeServices project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package requests

import (
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

func GetInt(r *http.Request, parameterName string) (int, bool) {
	if paramValue := GetSingleQueryParameter(r, parameterName, ""); paramValue == "" {
		return 0, false
	} else if intVal, err := strconv.Atoi(paramValue); err != nil {
		return 0, false
	} else {
		return intVal, true
	}
}

