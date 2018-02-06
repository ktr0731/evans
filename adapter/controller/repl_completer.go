package controller

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/entity"
)

type completer struct {
	cmds map[string]Commander
	env  *entity.Env
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
				{Text: "header"},
			}
		}

	case "package":
		pkgs := c.env.Packages()
		s = make([]prompt.Suggest, len(pkgs))
		for i, pkg := range pkgs {
			s[i] = prompt.Suggest{Text: pkg.Name}
		}

	case "service":
		svcs, err := c.env.Services()
		if err != nil {
			return nil
		}
		s = make([]prompt.Suggest, len(svcs))
		for i, svc := range svcs {
			s[i] = prompt.Suggest{Text: svc.Name()}
		}

	case "call":
		rpcs, err := c.env.RPCs()
		if err != nil {
			return nil
		}
		s = make([]prompt.Suggest, len(rpcs))
		for i, rpc := range rpcs {
			s[i] = prompt.Suggest{Text: rpc.Name()}
		}

	case "desc":
		msgs, err := c.env.Messages()
		if err != nil {
			return nil
		}
		s = make([]prompt.Suggest, len(msgs))
		for i, msg := range msgs {
			s[i] = prompt.Suggest{Text: msg.Name()}
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
