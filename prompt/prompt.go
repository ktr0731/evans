// Package prompt provides the prompt interface and its implementation.
package prompt

import (
	"fmt"
	"io"
	"os"

	"github.com/chzyer/readline"
	goprompt "github.com/ktr0731/go-prompt"
	"github.com/manifoldco/promptui"
	"github.com/pkg/errors"
)

type stdout struct{}

func (s *stdout) Write(b []byte) (int, error) {
	if len(b) == 1 && b[0] == 7 {
		return 0, nil
	}
	return os.Stdout.Write(b)
}

func (s *stdout) Close() error {
	return os.Stdout.Close()
}

func init() {
	// Override readline.Stdout to suppress ringing bell.
	// See more details: https://github.com/manifoldco/promptui/issues/49#issuecomment-428801411
	readline.Stdout = &stdout{}
}

// InitialColor is the initial color for a prompt prefix.
var (
	ColorInitial = Color(goprompt.DarkGreen)
	ColorBlue    = Color(goprompt.Blue)
)

var (
	ErrAbort = errors.New("abort")
)

// Color represents a valid color for a prompt prefix.
type Color goprompt.Color

// Next returns the next color of c. Note that Next will circular if c is the end of colors.
func (c *Color) Next() {
	*c = (*c + 1) % 16
}

// NextVal is the same as Next, but return color as value.
func (c *Color) NextVal() Color {
	return (*c + 1) % 16
}

type Prompt interface {
	// Input reads keyboard input.
	// If ctrl+d is entered, Input returns io.EOF.
	// If ctrl+c is entered, Input returns ErrAbort.
	Input() (string, error)
	Select(message string, options []string) (idx int, selected string, _ error)

	// SetPrefix changes the current prefix to the passed one.
	SetPrefix(prefix string)

	// SetPrefixColor changes the current color to the passed one.
	SetPrefixColor(color Color)

	// SetCompleter set a completer for prompt completion.
	SetCompleter(c Completer)

	// GetCommandHistory gets a command history. The order of history is asc.
	GetCommandHistory() []string
}

// New instantiates a new Prompt implementation. New will be replaced when e2egen command is executed.
// Initially, Prompt doesn't have a prefix, so you have to call SetPrefix for displaying it.
var New = newPrompt

func newPrompt(opts ...Option) Prompt {
	var opt opt
	for _, o := range opts {
		o(&opt)
	}

	p := &prompt{
		InputFunc:   goprompt.Input,
		prefixColor: ColorInitial,
		SelectFunc: func(message string, options []string) (int, string, error) {
			s := promptui.Select{
				Label:     message,
				Items:     options,
				Templates: &promptui.SelectTemplates{Label: fmt.Sprintf("%s {{.}}", promptui.IconInitial)},
			}
			return s.Run()
		},
		commandHistory: opt.commandHistory,
	}

	p.options = []goprompt.Option{
		goprompt.OptionLivePrefix(p.livePrefix),

		goprompt.OptionSuggestionBGColor(goprompt.LightGray),
		goprompt.OptionSuggestionTextColor(goprompt.Black),
		goprompt.OptionDescriptionBGColor(goprompt.White),
		goprompt.OptionDescriptionTextColor(goprompt.Black),

		goprompt.OptionSelectedSuggestionBGColor(goprompt.DarkBlue),
		goprompt.OptionSelectedSuggestionTextColor(goprompt.Black),
		goprompt.OptionSelectedDescriptionBGColor(goprompt.Blue),
		goprompt.OptionSelectedDescriptionTextColor(goprompt.Black),

		goprompt.OptionHistory(p.commandHistory),
	}
	return p
}

type prompt struct {
	prefix         string
	prefixColor    Color
	completer      Completer
	commandHistory []string
	options        []goprompt.Option

	// Treat prompt functions as fields for testing.
	InputFunc  func(prefix string, completer goprompt.Completer, opts ...goprompt.Option) (string, error)
	SelectFunc func(message string, options []string) (int, string, error)
}

func (p *prompt) Input() (in string, err error) {
	in, err = p.InputFunc(
		p.prefix,
		toGoPromptCompleter(p.completer),
		append(
			p.options,
			goprompt.OptionPrefixTextColor(goprompt.Color(p.prefixColor)),
			goprompt.OptionHistory(p.commandHistory),
		)...)
	if errors.Is(err, goprompt.ErrAbort) {
		return "", ErrAbort
	} else if err != nil {
		return "", err
	}
	p.commandHistory = append(p.commandHistory, in)
	return in, nil
}

func (p *prompt) Select(message string, options []string) (int, string, error) {
	n, res, err := p.SelectFunc(message, options)
	if errors.Is(err, promptui.ErrInterrupt) {
		return 0, "", ErrAbort
	}
	if errors.Is(err, promptui.ErrEOF) {
		return 0, "", io.EOF
	}
	if err != nil {
		return 0, "", errors.Wrap(err, "failed to select an item")
	}
	return n, res, nil
}

func (p *prompt) SetPrefix(prefix string) {
	p.prefix = prefix
}

func (p *prompt) SetPrefixColor(color Color) {
	p.prefixColor = color
}

func (p *prompt) SetCompleter(c Completer) {
	p.completer = c
}

func (p *prompt) GetCommandHistory() []string {
	return p.commandHistory
}

func (p *prompt) livePrefix() (string, bool) {
	return p.prefix, true
}

// Completer is a mechanism that provides REPL completion.
type Completer interface {
	// Complete receives d that is a piece of input, and returns some suggestions.
	// The prompt shows suggestions from it.
	Complete(d Document) []*Suggest
}

// Document is a piece of input that has several information such that
// text before the cursor, a work before the cursor, etc.
type Document interface {
	GetWordBeforeCursor() string
	TextBeforeCursor() string
}

// Suggest is a suggestion for the completion.
type Suggest struct {
	goprompt.Suggest
}

// NewSuggestion returns a new *Suggest from text and description.
func NewSuggestion(text, description string) *Suggest {
	return &Suggest{
		goprompt.Suggest{
			Text:        text,
			Description: description,
		},
	}
}

// FilterHasPrefix filters s by whether have sub as the prefix.
// If ignoreCase is true, differences between upper and lower casing are ignored.
func FilterHasPrefix(s []*Suggest, sub string, ignoreCase bool) []*Suggest {
	return fromGoPromptSuggestions(goprompt.FilterHasPrefix(fromPromptSuggestions(s), sub, ignoreCase))
}

func toGoPromptCompleter(c Completer) goprompt.Completer {
	if c == nil {
		return func(goprompt.Document) []goprompt.Suggest { return nil }
	}
	return func(d goprompt.Document) []goprompt.Suggest {
		s := c.Complete(&d)
		suggestions := make([]goprompt.Suggest, len(s))
		for i := 0; i < len(s); i++ {
			suggestions[i] = s[i].Suggest
		}
		return suggestions
	}
}

func fromGoPromptSuggestions(s []goprompt.Suggest) []*Suggest {
	suggestions := make([]*Suggest, len(s))
	for i, s := range s {
		suggestions[i] = NewSuggestion(s.Text, s.Description)
	}
	return suggestions
}

func fromPromptSuggestions(s []*Suggest) []goprompt.Suggest {
	suggestions := make([]goprompt.Suggest, len(s))
	for i, s := range s {
		suggestions[i] = goprompt.Suggest{Text: s.Text, Description: s.Description}
	}
	return suggestions
}
