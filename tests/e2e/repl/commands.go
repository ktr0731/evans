package repl

type CmdAndArgs func() []string

func Call(rpcName string, in ...string) CmdAndArgs {
	return func() []string {
		return append([]string{"call " + rpcName}, in...)
	}
}

func Help() []string {
	return []string{"help"}
}

func Exit() []string {
	return []string{"exit"}
}
