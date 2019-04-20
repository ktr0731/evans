package cui

import (
	"io"
)

// Option for basicUI
type Option func(*basicUI)

func Writer(w io.Writer) Option {
	return func(u *basicUI) {
		u.writer = w
	}
}

func ErrWriter(ew io.Writer) Option {
	return func(u *basicUI) {
		u.errWriter = ew
	}
}
