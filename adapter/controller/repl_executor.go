package controller

import (
	"strings"

	"github.com/k0kubun/pp"
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

	result, err := e.repl.eval(l)
	if err != nil {
		e.repl.ui.ErrPrintln(err.Error())
		return
	}

	pp.Println("returned")
	e.repl.ui.Println(result)
	e.repl.prompt.SetPrefix(e.repl.getPrompt())
}
