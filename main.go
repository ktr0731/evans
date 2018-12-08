package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/app"
)

func main() {
	os.Exit(app.New(nil).Run(os.Args[1:]))
}
