package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/cmd"
	"github.com/ktr0731/evans/adapter/cui"
	"github.com/ktr0731/evans/meta"
)

func main() {
	os.Exit(cmd.New(
		meta.AppName,
		meta.Version.String(),
		cui.NewBasic(),
	).Run(os.Args[1:]))
}
