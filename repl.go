package main

import (
	"bufio"
	"fmt"
)

var (
	CmdQuit = "quit"
	CmdExit = "exit"
)

type REPL struct {
	ui     *UI
	config *Config
}

type Config struct {
	Port int
}

func NewREPL(config *Config) *REPL {
	return &REPL{
		ui:     NewUI(),
		config: config,
	}
}

func (r *REPL) ShowPrompt() {
	fmt.Fprintf(r.ui.Writer, "localhost:%d> ", r.config.Port)
}

func (r *REPL) Response(text string) {
	fmt.Fprintf(r.ui.Writer, "%s\n", text)
}

func (r *REPL) Start() error {
	r.ShowPrompt()
	scanner := bufio.NewScanner(r.ui.Reader)

REPL:
	for scanner.Scan() {
		cmd := scanner.Text()

		switch cmd {
		case CmdQuit, CmdExit:
			r.Response("Bye!")
			break REPL
		default:
			r.Response(cmd)
		}

		r.ShowPrompt()
	}
	return scanner.Err()
}
