package repl

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

func completer(d prompt.Document) []prompt.Suggest {
	bc := d.TextBeforeCursor()
	if bc == "" {
		return nil
	}

	args := strings.Split(bc, " ")
	var s []prompt.Suggest
	switch args[0] {
	case "show":
		s = []prompt.Suggest{
			{Text: "package"},
			{Text: "service"},
			{Text: "message"},
			{Text: "rpc"},
		}
	default:
		// return all commands if current input is first command name
		if len(args) == 1 {
			s = []prompt.Suggest{
				{Text: "call"},
				{Text: "desc"},
				{Text: "package"},
				{Text: "service"},
				{Text: "show"},
			}
		}

	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
