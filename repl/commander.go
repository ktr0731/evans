package repl

type Commander interface {
	Help() string
	Validate(args []string) error
	Run(args []string) (string, error)
}
