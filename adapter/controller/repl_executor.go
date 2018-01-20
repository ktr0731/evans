package controller

import (
	"os"
	"strings"

	prompt "github.com/c-bata/go-prompt"
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
		err := prompt.OptionPrefix(e.repl.getPrompt())(e.repl.prompt)
		if err != nil {
			e.repl.wrappedError(err)
		}
	}
}
