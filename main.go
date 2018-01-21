package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
)

func main() {
	os.Exit(run(controller.NewCLI("evans", "0.1.0")))
}

func run(runnable controller.Runnable) int {
	return runnable.Run(os.Args[1:])
}
