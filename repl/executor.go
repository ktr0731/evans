package repl

import (
	"os"
	"strings"
)

func executor(r *REPL) func(string) {
	return func(l string) {
		if l == "quit" || l == "exit" {
			os.Exit(0)
			return
		}

		// Ignroe spaces
		if len(strings.TrimSpace(l)) == 0 {
			return
		}

		result, err := r.eval(l)
		if err != nil {
			r.wrappedError(err)
		} else {
			r.wrappedPrint(result)
		}
	}
}
