package cui

import (
	"fmt"
	"io"

	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

// UI provides formatted I/O interfaces.
// It is used from Evans's standard I/O and CLI mode I/O.
type UI interface {
	Output(s string)
	Info(s string)
	Error(s string)

	Writer() io.Writer
}

// New creates a new UI with passed options.
func New(opts ...Option) UI {
	// Creates a new UI with stdin, stdout, stderr.
	ui := &basicUI{
		writer:    colorable.NewColorableStdout(),
		errWriter: colorable.NewColorableStderr(),
	}
	for _, opt := range opts {
		opt(ui)
	}
	return ui
}

type basicUI struct {
	writer, errWriter io.Writer
}

// Output writes out the passed argument s to Writer with a line break.
func (u *basicUI) Output(s string) {
	fmt.Fprintln(u.writer, s)
}

// Info is the same as Output, but distinguish these for composition.
func (u *basicUI) Info(s string) {
	u.Output(s)
}

// Error writes out the passed argument s to ErrWriter with a line break.
func (u *basicUI) Error(s string) {
	fmt.Fprintln(u.errWriter, s)
}

// Writer returns an io.Writer which is used in u.
func (u *basicUI) Writer() io.Writer {
	return u.writer
}

type coloredUI struct {
	UI
}

// NewColored wraps provided `ui` with coloredUI.
// If `ui` is *coloredUI, NewColored returns it as it is.
// Colored output works fine in Windows environment.
func NewColored(ui UI) UI {
	if ui, ok := ui.(*coloredUI); ok {
		return ui
	}
	return &coloredUI{ui}
}

// Info is the same as New, but colored.
func (u *coloredUI) Info(s string) {
	u.UI.Info(color.BlueString(s))
}

// Error is the same as New, but colored.
func (u *coloredUI) Error(s string) {
	u.UI.Error(color.RedString(s))
}
