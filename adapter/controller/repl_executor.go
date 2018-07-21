package controller

import (
	"strings"
	"time"
)

type executor struct {
	repl *REPL
}

func (e *executor) execute(l string) {
	if l == "quit" || l == "exit" {
		e.repl.exitCh <- struct{}{}

		// do nothing, block execute method until Evans will be finished.
		//
		// if no sleep, c-bata/go-prompt will call Setup method within Run method.
		// then, tty's config is changed to raw mode.
		time.Sleep(10 * time.Minute)
	}

	// ignore spaces
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
