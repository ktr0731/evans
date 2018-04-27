package controller

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type UI interface {
	Println(s string)
	InfoPrintln(s string)
	ErrPrintln(s string)

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

func (u *BaseUI) Println(s string) {
	fmt.Fprintln(u.writer, s)
}

func (u *BaseUI) InfoPrintln(s string) {
	fmt.Fprintln(u.writer, s)
}

func (u *BaseUI) ErrPrintln(s string) {
	fmt.Fprintln(u.errWriter, s)
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

func (u *REPLUI) Println(s string) {
	u.UI.Println(s)
}

func (u *REPLUI) InfoPrintln(s string) {
	u.UI.InfoPrintln(s)
}

func (u *REPLUI) ErrPrintln(s string) {
	u.UI.ErrPrintln(s)
}

type ColoredUI struct {
	UI
}

func newColoredUI() *ColoredUI {
	return &ColoredUI{
		NewBasicUI(),
	}
}

func (u *ColoredUI) InfoPrintln(s string) {
	u.UI.InfoPrintln(color.BlueString(s))
}

func (u *ColoredUI) ErrPrintln(s string) {
	u.UI.ErrPrintln(color.RedString(s))
}

type ColoredREPLUI struct {
	*REPLUI
}

func newColoredREPLUI(prompt string) *ColoredREPLUI {
	return &ColoredREPLUI{
		newREPLUI(prompt),
	}
}

func (u *ColoredREPLUI) InfoPrintln(s string) {
	u.UI.InfoPrintln(color.BlueString(s))
}

func (u *ColoredREPLUI) ErrPrintln(s string) {
	u.UI.ErrPrintln(color.RedString(s))
}
