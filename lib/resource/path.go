package resource

import (
	"fmt"
	"strings"
)

// A standardized path is a path that starts with '/' and has no double slashes
type StandardizedPath string

/** transform a path to a standardized path
 * Watered down version of path.Clean. Replace any sequence of '/' with single '/'
 * Remove ending '/'
 * We do not resolve '..', (so '/foo/../baa' is different from '/baa')
 * Examples:
 *       '//foo/baa' becomes '/foo/baa'
 *       '/foo///baa/////muh/' becomes '/foo/baa/muh'
 *       '/foo/..//baa//' becomes '/foo/../baa'
 *       '/foo/baa' becomes (stays) '/foo/baa'
 */
func Standardize(p string) StandardizedPath {
	if len(p) == 0 || p[0] != '/' {
		panic(fmt.Sprintf("path must start with '/': '%s'", p))
	}

	var buffer = make([]byte, len(p), len(p))
	var pos = 0
	var justSawSlash = false

	for i := 0; i < len(p); i++ {
		if !justSawSlash || p[i] != '/' {
			buffer[pos] = p[i]
			pos++
		}
		justSawSlash = p[i] == '/'
	}

	if buffer[pos-1] == '/' {
		return StandardizedPath(buffer[:pos-1])
	} else {
		return StandardizedPath(buffer[:pos])
	}

}

func Standardizef(format string, args...interface{}) StandardizedPath {
	return Standardize(fmt.Sprintf(format, args...))
}

/**
	Break standardized path into dir-part and base-part
    '/foo/baa/res' -> '/foo/baa', 'res'
    '/foo/baa' -> '/foo', 'baa'
 */
func separate(sp StandardizedPath) (StandardizedPath, string) {
	if len(sp) == 0 {
		panic("Separating empty string")
	} else {
		var pos = strings.LastIndexByte(string(sp[:len(sp)-1]), '/')
		return sp[:pos], string(sp[pos+1:])
	}
}

