package controller

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type UI interface {
	Println(a interface{})
	InfoPrintln(a interface{})
	ErrPrintln(a interface{})

	Writer() io.Writer
	ErrWriter() io.Writer
}

type BaseUI struct {
	reader            io.Reader
	writer, errWriter io.Writer
}

func NewUI(r io.Reader, w, ew io.Writer) UI {
	return &BaseUI{
		reader:    r,
		writer:    w,
		errWriter: ew,
	}
}

func NewBasicUI() *BaseUI {
	return &BaseUI{
		reader:    os.Stdin,
		writer:    os.Stdout,
		errWriter: os.Stderr,
	}
}

func (u *BaseUI) fprintln(w io.Writer, a interface{}) {
	if i, ok := a.(io.Reader); ok {
		io.Copy(u.writer, i)
	} else {
		fmt.Fprintln(w, a)
	}
}

func (u *BaseUI) Println(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *BaseUI) InfoPrintln(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *BaseUI) ErrPrintln(a interface{}) {
	u.fprintln(u.errWriter, a)
}

func (u *BaseUI) Writer() io.Writer {
	return u.writer
}

func (u *BaseUI) ErrWriter() io.Writer {
	return u.errWriter
}

type REPLUI struct {
	UI
	prompt string
}

func newREPLUI(prompt string) *REPLUI {
	return &REPLUI{
		UI:     NewBasicUI(),
		prompt: prompt,
	}
}

func (u *REPLUI) Println(a interface{}) {
	u.UI.Println(a)
}

func (u *REPLUI) InfoPrintln(a interface{}) {
	u.UI.InfoPrintln(a)
}

func (u *REPLUI) ErrPrintln(a interface{}) {
	u.UI.ErrPrintln(a)
}

type ColoredUI struct {
	UI
}

func (u *ColoredUI) printWithColor(
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

func (u *ColoredUI) InfoPrintln(a interface{}) {
	u.printWithColor(u.UI.InfoPrintln, color.BlueString, a)
}

func (u *ColoredUI) ErrPrintln(a interface{}) {
	u.printWithColor(u.UI.ErrPrintln, color.RedString, a)
}

type ColoredREPLUI struct {
	*REPLUI
}

func newColoredREPLUI(ui *REPLUI) *ColoredREPLUI {
	ui.UI = &ColoredUI{ui.UI}
	return &ColoredREPLUI{ui}
}

func (u *ColoredREPLUI) InfoPrintln(a interface{}) {
	u.UI.InfoPrintln(a)
}

func (u *ColoredREPLUI) ErrPrintln(a interface{}) {
	u.UI.ErrPrintln(a)
}
