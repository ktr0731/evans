package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/meta"
)

func main() {
	os.Exit(controller.NewCLI(
		meta.AppName,
		meta.Version.String(),
	).Run(os.Args[1:]))
}
