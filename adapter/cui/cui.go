package cui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type CUI interface {
	Println(a interface{})
	InfoPrintln(a interface{})
	ErrPrintln(a interface{})

	Writer() io.Writer
	ErrWriter() io.Writer
}

type BaseCUI struct {
	reader            io.Reader
	writer, errWriter io.Writer
}

func New(r io.Reader, w, ew io.Writer) CUI {
	return &BaseCUI{
		reader:    r,
		writer:    w,
		errWriter: ew,
	}
}

func NewBasic() *BaseCUI {
	return &BaseCUI{
		reader:    os.Stdin,
		writer:    os.Stdout,
		errWriter: os.Stderr,
	}
}

func (u *BaseCUI) fprintln(w io.Writer, a interface{}) {
	if i, ok := a.(io.Reader); ok {
		io.Copy(u.writer, i)
	} else {
		fmt.Fprintln(w, a)
	}
}

func (u *BaseCUI) Println(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *BaseCUI) InfoPrintln(a interface{}) {
	u.fprintln(u.writer, a)
}

func (u *BaseCUI) ErrPrintln(a interface{}) {
	u.fprintln(u.errWriter, a)
}

func (u *BaseCUI) Writer() io.Writer {
	return u.writer
}

func (u *BaseCUI) ErrWriter() io.Writer {
	return u.errWriter
}

type ColoredCUI struct {
	CUI
}

func (u *ColoredCUI) printWithColor(
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

func (u *ColoredCUI) InfoPrintln(a interface{}) {
	u.printWithColor(u.CUI.InfoPrintln, color.BlueString, a)
}

func (u *ColoredCUI) ErrPrintln(a interface{}) {
	u.printWithColor(u.CUI.ErrPrintln, color.RedString, a)
}
