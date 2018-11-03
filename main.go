package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/cmd"
)

func main() {
	os.Exit(cmd.New(nil).Run(os.Args[1:]))
}
