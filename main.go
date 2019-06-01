package main

import (
	"os"

	"github.com/ktr0731/evans/app"
	"github.com/ktr0731/evans/cui"
)

func main() {
	os.Exit(app.New(cui.New()).Run(os.Args[1:]))
}
