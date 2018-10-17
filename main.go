package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/meta"
)

func main() {
	os.Exit(controller.NewCommand(
		meta.AppName,
		meta.Version.String(),
		cui.NewBasicUI(),
	).Run(os.Args[1:]))
}
