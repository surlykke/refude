package log

import (
	"fmt"
	golog "log"
	"os"
)

type LogLevel uint8

func makeLogger(level string) *golog.Logger {
	return golog.New(os.Stdout, level, golog.Ldate|golog.Ltime|golog.Lshortfile)
}

var debugLogger = makeLogger("DEBUG")
var infoLogger = makeLogger("INFO ")
var warnLogger = makeLogger("WARN ")
var errorLogger = makeLogger("ERROR")
var panicLogger = makeLogger("PANIC")

const level = 1 // 0: debug, 1: info, 2: warn, 3: error, 4: panic

func Debug(v ...interface{}) {
	if level == 0 {
		debugLogger.Output(3, fmt.Sprintln(v...))
	}
}

func Info(v ...interface{}) {
	if level <= 1 {
		infoLogger.Output(3, fmt.Sprintln(v...))
	}
}

func Warn(v ...interface{}) {
	if level <= 2 {
		warnLogger.Output(3, fmt.Sprintln(v...))
	}
}

func Error(v ...interface{}) {
	if level <= 3 {
		errorLogger.Output(3, fmt.Sprintln(v...))
	}
}

func Panic(v ...interface{}) {
	var s = fmt.Sprintln(v...)
	panicLogger.Output(3, fmt.Sprintln(v...))
	panic(s)
}
