package repl

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/k0kubun/pp"
	"github.com/ktr0731/evans/logger"
	"github.com/ktr0731/evans/prompt"
	"github.com/ktr0731/evans/usecase"
	"github.com/spf13/pflag"
)

var spaces = regexp.MustCompile(`\s+`)

type completer struct {
	cmds        map[string]commander
	completions map[string]func(args []string) (s []*prompt.Suggest)
}

// Complete completes suggestions from the input. In the completion, if an error is occurred, it will be ignored.
func (c *completer) Complete(d prompt.Document) (s []*prompt.Suggest) {
	bc := d.TextBeforeCursor()
	if bc == "" {
		return nil
	}

	// TODO: We should consider about spaces used as a part of test.
	args := strings.Split(spaces.ReplaceAllString(bc, " "), " ")

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

		for _, svc := range usecase.ListServicesOld() {
			s = append(s, prompt.NewSuggestion(svc, ""))
		}

	case "call":
		cmd := c.cmds["call"]
		fs, _ := cmd.FlagSet()

		if args[len(args)-1] == "-" || args[len(args)-1] == "--" {
			fs.VisitAll(func(f *pflag.Flag) {
				s = append(s, prompt.NewSuggestion("--"+f.Name, f.Usage))
			})
			return s
		}

		err := fs.Parse(args[1:]) // Ignore command name.
		if err != nil {
			logger.Printf("failed to parse flag: %s\n", err)
			return s
		}

		return c.completions["call"](fs.Args())
		// cmd := commands["call"].(*callCommand)
		// cmd.init() // TODO: constructor
		// _ = cmd.fs.Parse(args)
		//
		// args := cmd.fs.Args()
		//
		// if len(args) != 2 {
		// 	return nil
		// }
		// rpcs, err := usecase.ListRPCs("")
		// if err != nil {
		// 	return nil
		// }
		// for _, rpc := range rpcs {
		// 	s = append(s, prompt.NewSuggestion(rpc.Name, ""))
		// }

	case "desc":
		if len(args) != 2 {
			return nil
		}

		encountered := make(map[string]interface{})
		for _, svc := range usecase.ListServicesOld() {
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

func newCompleter(cmds map[string]commander) *completer {
	return &completer{
		cmds: cmds,
		completions: map[string]func(args []string) (s []*prompt.Suggest){
			"call": func(args []string) (s []*prompt.Suggest) {
				pp.Println(args)
				if len(args) == 1 && args[0] == "" {
					rpcs, err := usecase.ListRPCs("")
					if err != nil {
						return nil
					}
					for _, rpc := range rpcs {
						s = append(s, prompt.NewSuggestion(rpc.Name, ""))
					}
				}
				return s
			},
		},
	}
}
