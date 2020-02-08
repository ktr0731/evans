package repl

import (
	"fmt"
	"strings"

	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/usecase"
)

type completer struct {
	cmds map[string]commander
}

// Complete completes suggestions from the input. In the completion, if an error is occurred, it will be ignored.
func (c *completer) Complete(d prompt.Document) (s []*prompt.Suggest) {
	bc := d.TextBeforeCursor()
	if bc == "" {
		return nil
	}

	args := strings.Split(bc, " ")

	var isDefault bool
	defer func() {
		if !isDefault && len(s) != 0 {
			s = append(s, prompt.NewSuggestion("--help", "show command help message"))
		}
	}()

	switch args[0] {
	case "show":
		if len(args) != 2 {
			return nil
		}
		s = append(s,
			prompt.NewSuggestion("package", "show loaded package names"),
			prompt.NewSuggestion("service", "show loaded service names"),
			prompt.NewSuggestion("message", "show loaded messsage names"),
			prompt.NewSuggestion("rpc", "show RPC names belong to the current selected service"),
			prompt.NewSuggestion("header", "show headers which will be added to each request"),
		)

	case "package":
		if len(args) != 2 {
			return nil
		}
		pkgs := usecase.ListPackages()
		for _, pkg := range pkgs {
			if pkg == "" {
				s = append(s, prompt.NewSuggestion(`''`, "default for package name unspecified protos"))
			} else {
				s = append(s, prompt.NewSuggestion(pkg, ""))
			}
		}

	case "service":
		if len(args) != 2 {
			return nil
		}

		for _, svc := range usecase.ListServices() {
			s = append(s, prompt.NewSuggestion(svc, ""))
		}

	case "call":
		if len(args) != 2 {
			return nil
		}
		rpcs, err := usecase.ListRPCs("")
		if err != nil {
			return nil
		}
		for _, rpc := range rpcs {
			s = append(s, prompt.NewSuggestion(rpc.Name, ""))
		}

	case "desc":
		if len(args) != 2 {
			return nil
		}

		encountered := make(map[string]interface{})
		for _, svc := range usecase.ListServices() {
			rpcs, err := usecase.ListRPCs(svc)
			if err != nil {
				panic(fmt.Sprintf("ListRPCs must not return an error, but got '%s'", err))
			}
			for _, rpc := range rpcs {
				if _, found := encountered[rpc.RequestType.Name]; !found {
					s = append(s, prompt.NewSuggestion(rpc.RequestType.Name, ""))
					encountered[rpc.RequestType.Name] = nil
				}
				if _, found := encountered[rpc.ResponseType.Name]; !found {
					s = append(s, prompt.NewSuggestion(rpc.ResponseType.Name, ""))
					encountered[rpc.ResponseType.Name] = nil
				}
			}
		}

	default:
		isDefault = true
		// return all commands if current input is first command name
		if len(args) == 1 {
			// number of commands + help
			cmdNames := make([]*prompt.Suggest, 0, len(c.cmds))
			cmdNames = append(cmdNames, prompt.NewSuggestion("help", "show help message"))
			for name, cmd := range c.cmds {
				cmdNames = append(cmdNames, prompt.NewSuggestion(name, cmd.Synopsis()))
			}

			s = cmdNames
		}
	}
	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}
