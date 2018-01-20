package controller

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
)

type ui interface {
	Println(s string)
	InfoPrintln(s string)
	ErrPrintln(s string)

	Writer() io.Writer
	ErrWriter() io.Writer
}

type UI struct {
	reader            io.Reader
	writer, errWriter io.Writer
}

func newUI() *UI {
	return &UI{
		reader:    os.Stdin,
		writer:    os.Stdout,
		errWriter: os.Stderr,
	}
}

func (u *UI) Println(s string) {
	fmt.Fprintln(u.writer, s)
}

func (u *UI) InfoPrintln(s string) {
	fmt.Fprintln(u.writer, s)
}

func (u *UI) ErrPrintln(s string) {
	fmt.Fprintln(u.errWriter, s)
}

func (u *UI) Writer() io.Writer {
	return u.writer
}

func (u *UI) ErrWriter() io.Writer {
	return u.errWriter
}

type REPLUI struct {
	ui
	prompt string
}

func newREPLUI(prompt string) *REPLUI {
	return &REPLUI{
		ui:     newUI(),
		prompt: prompt,
	}
}

func (u *REPLUI) Println(s string) {
	u.ui.Println(s)
}

func (u *REPLUI) InfoPrintln(s string) {
	u.ui.InfoPrintln(s)
}

func (u *REPLUI) ErrPrintln(s string) {
	u.ui.ErrPrintln(s)
}

type ColoredUI struct {
	ui
}

func newColoredUI() *ColoredUI {
	return &ColoredUI{
		newUI(),
	}
}

func (u *ColoredUI) InfoPrintln(s string) {
	u.ui.InfoPrintln(color.BlueString(s))
}

func (u *ColoredUI) ErrPrintln(s string) {
	u.ui.ErrPrintln(color.RedString(s))
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
	u.ui.InfoPrintln(color.BlueString(s))
}

func (u *ColoredREPLUI) ErrPrintln(s string) {
	u.ui.ErrPrintln(color.RedString(s))
}
