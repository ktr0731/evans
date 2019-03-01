// Package logger provides logging functions.
// As default, logger discards all passed messages. See SetOutput for more details.
package logger

import (
	"io"
	"io/ioutil"
	"log"
)

var (
	defaultLogger = newDefaultLogger()
	enabled       bool
)

// Reset resets logging all parameters.
func Reset() {
	defaultLogger = newDefaultLogger()
	enabled = false
}

// SetOutput enables logging that writes out logs to w.
func SetOutput(w io.Writer) {
	enabled = true
	defaultLogger.SetOutput(w)
}

// SetPrefix changes the log prefix to another one.
// The default prefix is "evans: ".
func SetPrefix(p string) {
	defaultLogger.SetPrefix(p)
}

func Println(v ...interface{}) {
	defaultLogger.Println(v...)
}

func Printf(format string, v ...interface{}) {
	defaultLogger.Printf(format, v...)
}

func Fatal(v ...interface{}) {
	defaultLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	defaultLogger.Fatalf(format, v...)
}

// Scriptln receives a function f which executes something and returns some values
// as a slice of empty interfaces. If logging is disabled, f is not executed.
func Scriptln(f func() []interface{}) {
	if !enabled {
		return
	}
	args := f()
	Println(args...)
}

func newDefaultLogger() *log.Logger {
	return log.New(ioutil.Discard, "evans: ", 0)
}
