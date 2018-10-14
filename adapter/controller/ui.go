package controller

import "github.com/ktr0731/evans/adapter/cui"

type REPLUI struct {
	cui.UI
	prompt string
}

func newREPLUI(prompt string) *REPLUI {
	return &REPLUI{
		UI:     cui.NewBasicUI(),
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

type ColoredREPLUI struct {
	*REPLUI
}

func newColoredREPLUI(ui *REPLUI) *ColoredREPLUI {
	ui.UI = &cui.ColoredUI{ui.UI}
	return &ColoredREPLUI{ui}
}

func (u *ColoredREPLUI) InfoPrintln(a interface{}) {
	u.UI.InfoPrintln(a)
}

func (u *ColoredREPLUI) ErrPrintln(a interface{}) {
	u.UI.ErrPrintln(a)
}
