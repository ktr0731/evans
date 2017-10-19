package main

import (
	"os"

	"github.com/ktr0731/evans/cli"
)

func main() {
	os.Exit(cli.NewCLI("evans", "0.1.0").Run(os.Args[1:]))
}
