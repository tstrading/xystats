package logger

import (
	"fmt"
	"log"
	"os"
	"reflect"
	"unsafe"
)

type Logger log.Logger

func getLogger(prefix string) *log.Logger {
	isProduction := os.Getenv("LOG_SHORT") != ""
	if os.Getenv("LOG_PREFIX") != "" {
		prefix = os.Getenv("LOG_PREFIX") + " " + prefix
	}
	if isProduction {
		return log.New(os.Stdout, prefix, log.Lshortfile)
	} else {
		return log.New(os.Stdout, prefix, log.Lshortfile|log.LstdFlags|log.Lmicroseconds)
	}
}

func unsafeBytesToString(b []byte) (s string) {
	var length = len(b)
	if length == 0 {
		return ""
	}
	stringHeader := (*reflect.StringHeader)(unsafe.Pointer(&s))
	stringHeader.Data = uintptr(unsafe.Pointer(&b[0]))
	stringHeader.Len = length
	return s
}

var debug = getLogger("D ")
var info = getLogger("I ")
var warn = getLogger("W ")
var err = getLogger("E ")
var logLevel = os.Getenv("LOG_LEVEL")

func Debugf(format string, v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" {
		debug.Output(2, "\t"+fmt.Sprintf(format, v...))
	}
}

func Infof(format string, v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" || logLevel == "INFO" {
		info.Output(2, "\t"+fmt.Sprintf(format, v...))
	}
}

func Warnf(format string, v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" || logLevel == "INFO" || logLevel == "WARN" {
		warn.Output(2, "\t"+fmt.Sprintf(format, v...))
	}
}

func Errorf(format string, v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" || logLevel == "INFO" || logLevel == "WARN" || logLevel == "ERROR" {
		warn.Output(2, "\t"+fmt.Sprintf(format, v...))
	}
}

func Fatalf(format string, v ...interface{}) {
	err.Output(2, "\t"+fmt.Sprintf(format, v...))
	os.Exit(1)
}

func DebugBytes(msg []byte) {
	if logLevel == "" || logLevel == "DEBUG" {
		debug.Output(2, unsafeBytesToString(msg))
	}
}

func Debug(v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" {
		debug.Output(2, "\t"+fmt.Sprintln(v...))
	}
}

func Info(v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" || logLevel == "INFO" {
		info.Output(2, "\t"+fmt.Sprintln(v...))
	}
}
func Warn(v ...interface{}) {
	if logLevel == "" || logLevel == "DEBUG" || logLevel == "INFO" || logLevel == "WARN" {
		warn.Output(2, "\t"+fmt.Sprintln(v...))
	}
}

func Fatal(v ...interface{}) {
	err.Output(2, "\t"+fmt.Sprintln(v...))
	os.Exit(1)
}

func Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	err.Output(2, "\t"+s)
	panic(s)
}

func Panic(v ...interface{}) {
	s := fmt.Sprintln(v...)
	err.Output(2, "\t"+s)
	panic(s)
}
