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

type logLevel uint8

const (
	debug logLevel = iota
	info
	warn
	error_
	fatal
	panic_
)

var BaseLevel logLevel = info

func (l logLevel) String() string {
	switch l {
	case debug:
		return "DEBUG"
	case info:
		return "INFO"
	case warn:
		return "WARN"
	case error_:
		return "ERROR"
	case fatal:
		return "FATAL"
	case panic_:
		return "PANIC"
	default:
		return ""
	}
}

func Debug(v ...any) {
	writeMsg(debug, v)
}

func Info(v ...any) {
	writeMsg(info, v)
}

func Warn(v ...any) {
	writeMsg(warn, v)
}

func Error(v ...any) {
	writeMsg(error_, v)
}

func Fatal(v ...any) {
	writeMsg(fatal, v)
	os.Exit(1)
}

func Panic(v ...any) {
	writeMsg(panic_, v)
	panic("")
}

func writeMsg(level logLevel, v []any) {
	if level >= BaseLevel {
		fmt.Fprint(os.Stderr, level, ": ")
		fmt.Fprintln(os.Stderr, v...)
	}
}
