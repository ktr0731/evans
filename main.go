package main

import (
	"os"

	"github.com/ktr0731/evans/app"
)

func main() {
	os.Exit(app.New(nil).Run(os.Args[1:]))
}
