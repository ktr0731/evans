package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/usecase/port"
)

func main() {
	os.Exit(run(controller.NewCLI("evans", "0.1.0")))
}

func run(runnable port.InputPort) int {
	return runnable.Run(os.Args[1:])
}
