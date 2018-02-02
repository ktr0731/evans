package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
)

const (
	name    = "evans"
	version = "0.1.2"
)

func main() {
	os.Exit(run(controller.NewCLI(name, version)))
}

func run(runnable controller.Runnable) int {
	return runnable.Run(os.Args[1:])
}
