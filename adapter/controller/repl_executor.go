package controller

import (
	"strings"
)

type executor struct {
	repl *REPL
}

func (e *executor) execute(l string) {
	if l == "quit" || l == "exit" {
		e.repl.exitCh <- struct{}{}
		return
	}

	// Ignroe spaces
	if len(strings.TrimSpace(l)) == 0 {
		return
	}

	// break one line
	defer e.repl.ui.Println("")

	result, err := e.repl.eval(l)
	if err != nil {
		e.repl.ui.ErrPrintln(err.Error())
		return
	}

	e.repl.ui.Println(result)
	e.repl.prompt.SetPrefix(e.repl.getPrompt())
}
