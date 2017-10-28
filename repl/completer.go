package repl

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/env"
)

type completer struct {
	cmds map[string]Commander
	env  *env.Env
}

func (c *completer) complete(d prompt.Document) []prompt.Suggest {
	bc := d.TextBeforeCursor()
	if bc == "" {
		return nil
	}

	args := strings.Split(bc, " ")

	var s []prompt.Suggest
	switch args[0] {
	case "show":
		if len(args) == 2 {
			s = []prompt.Suggest{
				{Text: "package"},
				{Text: "service"},
				{Text: "message"},
				{Text: "rpc"},
			}
		}
	default:
		// return all commands if current input is first command name
		if len(args) == 1 {
			// number of commands + help
			cmdNames := make([]prompt.Suggest, len(c.cmds)+1)
			cmdNames = append(cmdNames, prompt.Suggest{Text: "help", Description: "show help message"})
			for name, cmd := range c.cmds {
				cmdNames = append(cmdNames, prompt.Suggest{Text: name, Description: cmd.Synopsis()})
			}

			s = cmdNames
		}

	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
