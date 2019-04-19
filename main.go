package main

import (
	"os"

	"github.com/ktr0731/evans/adapter/app"
	"github.com/ktr0731/evans/adapter/cui"
)

func main() {
	os.Exit(app.New(cui.New()).Run(os.Args[1:]))
}
