// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package log

import (
	"fmt"
	"os"
)

const loglevel uint8 = 0 // 0 warn, 1 error, >= 2 panic

func Warn(v ...any) {
	if loglevel < 1 {
		fmt.Fprintln(os.Stderr, v...)
	}
}

func Error(v ...any) {
	if loglevel < 2 {
		fmt.Fprintln(os.Stderr, v...)
	}
}

func Panic(v ...any) {
	var s = fmt.Sprintln(v...)
	fmt.Fprintln(os.Stderr, s)
	panic(s)
}
