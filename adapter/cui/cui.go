package cui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	colorable "github.com/mattn/go-colorable"
)

// UI provides formatted I/O interfaces.
// It is used from Evans's standard I/O and CLI mode I/O.
type UI interface {
	Println(a interface{})
	InfoPrintln(a interface{})
	ErrPrintln(a interface{})

	Writer() io.Writer
	ErrWriter() io.Writer
}

type basicUI struct {
	reader            io.Reader
	writer, errWriter io.Writer
}

// New creates a new UI with passed io.Reader, io.Writers.
// In normal case, you can use NewBasic instead of New.
func New(r io.Reader, w, ew io.Writer) UI {
	return &basicUI{
		reader:    r,
		writer:    w,
		errWriter: ew,
	}
}

// NewBasic creates a new UI with stdin, stdout, stderr.
func NewBasic() UI {
	return &basicUI{
		reader:    os.Stdin,
		writer:    colorable.NewColorableStdout(),
		errWriter: colorable.NewColorableStderr(),
	}
}

func (u *basicUI) fprintln(w io.Writer, a interface{}) {
	if i, ok := a.(io.Reader); ok {
		io.Copy(u.writer, i)
	} else {
		fmt.Fprintln(w, a)
	}
}

func (u *basicUI) Println(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *basicUI) InfoPrintln(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *basicUI) ErrPrintln(a interface{}) {
	u.fprintln(u.errWriter, a)
}

func (u *basicUI) Writer() io.Writer {
	return u.writer
}

func (u *basicUI) ErrWriter() io.Writer {
	return u.errWriter
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

func (u *coloredUI) printWithColor(
	w func(a interface{}),
	color func(format string, a ...interface{}) string,
	a interface{},
) {
	switch t := a.(type) {
	case string:
		w(color(t))
	case fmt.Stringer:
		w(color(t.String()))
	default:
		w(t)
	}
}

func (u *coloredUI) InfoPrintln(a interface{}) {
	u.printWithColor(u.UI.InfoPrintln, color.BlueString, a)
}

func (u *coloredUI) ErrPrintln(a interface{}) {
	u.printWithColor(u.UI.ErrPrintln, color.RedString, a)
}
