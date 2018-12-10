package logger

import (
	"io"
	"log"
	"os"
)

var (
	defaultLogger = log.New(os.Stderr, "evans: ", 0)
)

func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
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
