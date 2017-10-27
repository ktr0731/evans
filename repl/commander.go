package repl

type Commander interface {
	Help() string
	Synopsis() string
	Validate(args []string) error
	Run(args []string) (string, error)
}
