package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/usecase"
)

func main() {
	interactor := usecase.NewInteractor()
	os.Exit(run(controller.NewCLI("evans", "0.1.0", interactor)))
}

func run(runnable controller.Runnable) int {
	return runnable.Run(os.Args[1:])
}
