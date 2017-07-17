package main

import (
	"os"
)

func main() {
	os.Exit(NewCLI("evans", "0.1.0").Run(os.Args[1:]))
}
