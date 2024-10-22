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
