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
