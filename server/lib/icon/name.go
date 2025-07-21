// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
//
package icon

import (
	"fmt"
	"strings"
)

type Name string

func (this Name) String() string {
	if this == "" {
		return ""
	} else if strings.HasPrefix(string(this), "http") {
		return string(this)
	} else {
		return fmt.Sprintf("/icon?name=%s", string(this))
	}
}
