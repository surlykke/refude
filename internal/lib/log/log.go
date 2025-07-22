// Copyright (c) Christian Surlykke
//
// This file is part of the refude project.
// It is distributed under the GPL v2 license.
// Please refer to the GPL2 file for a copy of the license.
package log

import (
	"fmt"
)

const loglevel = 1 // 0: debug, 1: info, 2: warn, 3: error, 4: panic

func Debug(v ...interface{}) {
	if 0 >= loglevel {
		fmt.Println(v...)
	}
}

func Info(v ...interface{}) {
	if 1 >= loglevel {
		fmt.Println(v...)
	}
}

func Warn(v ...interface{}) {
	if 1 >= loglevel {
		fmt.Println(v...)
	}
}

func Error(v ...interface{}) {
	if 1 >= loglevel {
		fmt.Println(v...)
	}
}

func Panic(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	fmt.Print(s)
	panic(s)
}
