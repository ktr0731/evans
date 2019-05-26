// Package cui defines charcter user interfaces for I/O.
package cui

import (
	"fmt"
	"io"
	"os"
)

var defaultUI = UI{
	Prefix:    "evans: ",
	Writer:    os.Stdout,
	ErrWriter: os.Stderr,
}

// UI provides formatted output for the application. Almost users can use
// DefaultUI() as it is.
type UI struct {
	Prefix            string
	Writer, ErrWriter io.Writer
}

// Output writes out the passed argument s to Writer with a line break.
func (u *UI) Output(s string) {
	fmt.Fprintln(u.Writer, s)
}

// Info is the same as Output, but distinguish these for composition.
func (u *UI) Info(s string) {
	u.Output(s)
}

// Error writes out the passed argument s to ErrWriter with a line break.
func (u *UI) Error(s string) {
	fmt.Fprintln(u.ErrWriter, s)
}

// DefaultUI returns the default UI.
func DefaultUI() *UI {
	ui := defaultUI
	return &ui
}
