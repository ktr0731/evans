package color

import prompt "github.com/c-bata/go-prompt"

type Color prompt.Color

func DefaultColor() Color {
	return Color(prompt.DarkGreen)
}

func (c *Color) Next() {
	*c = (*c + 1) % 16
}

func (c *Color) Prev() {
	*c = (*c - 1) % 16
}
