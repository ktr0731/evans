package prompt

type opt struct {
	commandHistory []string
}

type Option func(*opt)

func WithCommandHistory(h []string) Option {
	return func(o *opt) {
		o.commandHistory = h
	}
}
