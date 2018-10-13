package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/cmd"
	"github.com/ktr0731/evans/adapter/controller"
	"github.com/ktr0731/evans/meta"
)

func main() {
	os.Exit(cmd.New(
		meta.AppName,
		meta.Version.String(),
		controller.NewBasicUI(),
	).Run(os.Args[1:]))
}
