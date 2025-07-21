// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package path

import (
	"fmt"
	"strings"
)

type Path string

func ToPath(path string) Path {
	if !strings.HasPrefix(path, "/") {
		panic("Paths must start with '/'")
	}

	return Path(path)
}

func Of(elements ...any) Path {
	var tmp = fmt.Sprint(elements...)
	if !strings.HasPrefix(tmp, "/") {
		panic("Paths must start with '/'")
	}

	return Path(tmp)
}
