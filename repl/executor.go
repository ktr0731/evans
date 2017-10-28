package repl

import (
	"os"
	"strings"
)

type executor struct {
	repl *REPL
}

func (e *executor) execute(l string) {
	if l == "quit" || l == "exit" {
		os.Exit(0)
		return
	}

	// Ignroe spaces
	if len(strings.TrimSpace(l)) == 0 {
		return
	}

	result, err := e.repl.eval(l)
	if err != nil {
		e.repl.wrappedError(err)
	} else {
		e.repl.wrappedPrint(result)
	}
}
