package prompt

import (
	"io"

	"github.com/AlecAivazis/survey"
	goprompt "github.com/c-bata/go-prompt"
	"github.com/ktr0731/evans/color"
)

// Prompt provides interactive interfaces to receive user input.
type Prompt interface {
	// Run executes Input continually.
	// It is called from REPL input prompter.
	// Run will be finished when a user enters CTRL+d.
	Run()

	// Input receives user entered input.
	// Input will be abort when a user enters CTRL+d.
	Input() (string, error)

	// Select displays a selection consists of opts.
	Select(msg string, opts []string) (string, error)

	// SetPrefix chnages current prompt prefix by passed one.
	SetPrefix(prefix string)

	// SetPrefixColor changes current prompt color by passed one.
	SetPrefixColor(color color.Color) error
}

type prompt struct {
	fieldPrompt   *goprompt.Prompt
	currentPrefix string

	hasExecutor bool

	// entered is changed from prompt.Enter key bind of c-bata/go-prompt.
	// c-bata/go-prompt doesn't return EOF (ctrl+d), only returns empty string.
	// So Evans cannot determine whether empty input is EOF or just entered.
	// Therefore, Evans uses a tricky method to know EOF.
	//
	// 1. Register key bind for enter at New.
	// 2. The key bind changes entered variable from  KeyBindFunc.
	// 3. prompt.Input checks entered field.
	//    If the input is empty string and entered field is false, it is EOF.
	// 4. Input returns io.EOF as an error.
	entered bool
}

// New instantiates a prompt which satisfied Prompt with c-bata/go-prompt.
// New will be replace by a mock when e2e testing.
//
// Prompt will panic when called Run if executor is nil.
//
// TODO: Don't declare New as a variable.
var New = func(executor func(string), completer func(goprompt.Document) []goprompt.Suggest, opt ...goprompt.Option) Prompt {
	return newPrompt(executor, completer, opt...)
}

func newPrompt(executor func(string), completer func(goprompt.Document) []goprompt.Suggest, opt ...goprompt.Option) Prompt {
	if executor == nil {
		executor = func(in string) {
			return
		}
	}
	if completer == nil {
		completer = func(d goprompt.Document) []goprompt.Suggest {
			return nil
		}
	}
	p := &prompt{
		hasExecutor: executor != nil,
	}
	p.fieldPrompt = goprompt.New(
		executor,
		completer,
		append(opt,
			goprompt.OptionLivePrefix(p.livePrefix),
			goprompt.OptionAddKeyBind(goprompt.KeyBind{
				Key: goprompt.Enter,
				Fn: func(_ *goprompt.Buffer) {
					p.entered = true
				},
			}),
		)...)
	return p
}

// Run receives user input and call executor function passed at New.
// If executor is nil, Run will panic.
func (p *prompt) Run() {
	if !p.hasExecutor {
		panic("prompt.Run requires executor function at New")
	}
	p.fieldPrompt.Run()
}

func (p *prompt) Input() (string, error) {
	p.entered = false
	in := p.fieldPrompt.Input()
	// ctrl+d
	if !p.entered && in == "" {
		return "", io.EOF
	}
	return in, nil
}

func (p *prompt) Select(msg string, opts []string) (string, error) {
	var choice string
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: opts,
	}, &choice, nil)
	if err != nil {
		return "", io.EOF
	}
	return choice, nil
}

func (p *prompt) SetPrefix(prefix string) {
	p.currentPrefix = prefix
}

func (p *prompt) SetPrefixColor(color color.Color) error {
	return goprompt.OptionPrefixTextColor(goprompt.Color(color))(p.fieldPrompt)
}

func (p *prompt) livePrefix() (string, bool) {
	return p.currentPrefix, true
}
