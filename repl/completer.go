package repl

import (
	"fmt"
	"regexp"
	"strings"

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
	cmdName := args[0]
	args = args[1:] // Ignore command name.

	cmd, ok := c.cmds[cmdName]
	if !ok {
		// return all commands if current input is first command name
		if len(args) == 0 {
			// number of commands + help
			cmdNames := make([]*prompt.Suggest, 0, len(c.cmds))
			cmdNames = append(cmdNames, prompt.NewSuggestion("help", "show help message"))
			for name, cmd := range c.cmds {
				cmdNames = append(cmdNames, prompt.NewSuggestion(name, cmd.Synopsis()))
			}

			s = cmdNames
		}
		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}

	defer func() {
		if len(s) != 0 {
			s = append(s, prompt.NewSuggestion("--help", "show command help message"))
		}
		s = prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}()

	fs, ok := cmd.FlagSet()
	if ok {
		if len(args) > 0 && strings.HasPrefix(args[len(args)-1], "-") {
			fs.VisitAll(func(f *pflag.Flag) {
				s = append(s, prompt.NewSuggestion("--"+f.Name, f.Usage))
			})
			return s
		}

		_ = fs.Parse(args)
		args = fs.Args()
	}

	compFunc, ok := c.completions[cmdName]
	if !ok {
		return s
	}
	return compFunc(args)
}

func newCompleter(cmds map[string]commander) *completer {
	return &completer{
		cmds: cmds,
		completions: map[string]func(args []string) (s []*prompt.Suggest){
			"show": func(args []string) (s []*prompt.Suggest) {
				if len(args) == 1 {
					s = []*prompt.Suggest{
						prompt.NewSuggestion("package", "show loaded package names"),
						prompt.NewSuggestion("service", "show loaded service names"),
						prompt.NewSuggestion("message", "show loaded message names"),
						prompt.NewSuggestion("rpc", "show RPC names belonging to the current selected service"),
						prompt.NewSuggestion("header", "show headers which will be added to each request"),
					}
				}
				return s
			},
			"package": func(args []string) (s []*prompt.Suggest) {
				if len(args) == 1 {
					pkgs := usecase.ListPackages()
					for _, pkg := range pkgs {
						if pkg == "" {
							s = append(s, prompt.NewSuggestion(`''`, "default for package name unspecified protos"))
						} else {
							s = append(s, prompt.NewSuggestion(pkg, ""))
						}
					}
				}
				return s
			},
			"service": func(args []string) (s []*prompt.Suggest) {
				if len(args) == 1 {
					for _, svc := range usecase.ListServices() {
						s = append(s, prompt.NewSuggestion(svc, ""))
					}
				}
				return s
			},
			"call": func(args []string) (s []*prompt.Suggest) {
				if len(args) == 1 {
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
			"desc": func(args []string) (s []*prompt.Suggest) {
				if len(args) != 1 {
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
				return s
			},
		},
	}
}
