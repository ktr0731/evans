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

// Reset resets all logging parameters.
func Reset() {
	defaultLogger = newDefaultLogger()
	enabled = false
}

// SetOutput enables logging that writes out logs to w.
// Note that SetOutput works only once. To perform SetOutput again,
// it is neccessary to call Reset before it.
func SetOutput(w io.Writer) {
	if enabled {
		Println("logger: ignored SetOutput because it is already called. please call Reset before calling again.")
		return
	}
	enabled = true
	defaultLogger.SetOutput(w)
}

// Println provides fmt.Println like logging.
func Println(v ...interface{}) {
	defaultLogger.Println(v...)
}

// Printf provides fmt.Printf like logging.
func Printf(format string, v ...interface{}) {
	defaultLogger.Printf(format, v...)
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

// Scriptf is similar with Scriptln, but for formatting output.
func Scriptf(format string, f func() []interface{}) {
	if !enabled {
		return
	}
	args := f()
	Printf(format, args...)
}

func newDefaultLogger() *log.Logger {
	return log.New(ioutil.Discard, "evans: ", 0)
}
