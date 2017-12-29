package port

type InputPort interface {
	Run(args []string) int
}
