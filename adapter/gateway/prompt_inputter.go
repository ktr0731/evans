package gateway

import (
	"io"
	"strings"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	shellstring "github.com/ktr0731/go-shellstring"
	"github.com/pkg/errors"
)

var (
	ErrUnknownOneofFieldName = errors.New("unknown oneof field name")
	ErrUnknownEnumName       = errors.New("unknown enum name")
	EORF                     = errors.New("end of repeated field")
)

// for mocking
type Prompter interface {
	// Run is called from REPL input prompter
	Run()

	Input() (string, error)
	Select(msg string, opts []string) (string, error)
	SetPrefix(prefix string)
	SetPrefixColor(color prompt.Color) error
}

type RealPrompter struct {
	fieldPrompter *prompt.Prompt
	currentPrefix string

	// entered is changed from prompt.Enter key bind of c-bata/go-prompt.
	// c-bata/go-prompt doesn't return EOF (ctrl+d), only returns empty string.
	// so Evans cannot determine whether empty input is EOF or just entered.
	// therefore, Evans uses a tricky method to know EOF.
	//
	// 1. register key bind for enter at NewRealPrompter.
	// 2. the key bind changes entered variable from  KeyBindFunc.
	// 3. RealPrompter.Input checks entered field.
	//    if the input is empty string and entered field is false, it is EOF.
	// 4. Input returns io.EOF as a error.
	entered bool
}

// NewRealPrompter instantiates a prompt which satisfied Prompter with go-prompt.
// NewRealPrompter will be replace by a mock when e2e testing.
//
// NewRealPrompter is called to create REPL-CLI and REPL field inputter.
// NewPrompt is the short-hand method to create *Prompt with no params NewRealPrompter.
var NewRealPrompter = func(executor func(string), completer func(prompt.Document) []prompt.Suggest, opt ...prompt.Option) Prompter {
	if executor == nil {
		executor = func(in string) {
			return
		}
	}
	if completer == nil {
		completer = func(d prompt.Document) []prompt.Suggest {
			return nil
		}
	}
	p := &RealPrompter{}
	p.fieldPrompter = prompt.New(
		executor,
		completer,
		append(opt,
			prompt.OptionLivePrefix(p.livePrefix),
			prompt.OptionAddKeyBind(prompt.KeyBind{
				Key: prompt.Enter,
				Fn: func(_ *prompt.Buffer) {
					p.entered = true
				},
			}),
		)...)
	return p
}

func (p *RealPrompter) Run() {
	p.fieldPrompter.Run()
}

func (p *RealPrompter) Input() (string, error) {
	p.entered = false
	in := p.fieldPrompter.Input()
	// ctrl+d
	if !p.entered && in == "" {
		return "", io.EOF
	}
	return in, nil
}

func (p *RealPrompter) Select(msg string, opts []string) (string, error) {
	var choice string
	err := survey.AskOne(&survey.Select{
		Message: msg,
		Options: opts,
	}, &choice, nil)
	return choice, err
}

func (p *RealPrompter) SetPrefix(prefix string) {
	p.currentPrefix = prefix
}

func (p *RealPrompter) SetPrefixColor(color prompt.Color) error {
	return prompt.OptionPrefixTextColor(color)(p.fieldPrompter)
}

func (p *RealPrompter) livePrefix() (string, bool) {
	return p.currentPrefix, true
}

// NewPrompt instantiates new *Prompt with NewRealPrompter.
func NewPrompt(config *config.Config, env entity.Environment) *Prompt {
	return newPrompt(NewRealPrompter(nil, nil), config, env)
}

// Prompt is an implementation of inputting method.
// it has common logic to input fields interactively.
// in normal, go-prompt is used as prompt.
type Prompt struct {
	prompt Prompter
	config *config.Config
	env    entity.Environment
}

func newPrompt(prompt Prompter, config *config.Config, env entity.Environment) *Prompt {
	return &Prompt{
		prompt: prompt,
		config: config,
		env:    env,
	}
}

// impl of port.Inputter
func (i *Prompt) Input(reqType entity.Message) (proto.Message, error) {
	setter := protobuf.NewMessageSetter(reqType)
	fields := reqType.Fields()

	// DarkGreen is the initial color
	return newFieldInputter(i.prompt, i.config.Input.PromptFormat, setter, []string{}, false, prompt.DarkGreen).Input(fields)
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
type fieldInputter struct {
	prompt Prompter
	setter *protobuf.MessageSetter

	prefixFormat string
	ancestor     []string

	color prompt.Color

	// enteredEmptyInput is used to terminate repeated field inputting
	// if input is empty and enteredEmptyInput is true, exit repeated input prompt
	enteredEmptyInput bool

	// the field has parent field (in other words, the field is child of a message field)
	hasAncestorAndHasRepeatedField bool
}

func newFieldInputter(
	prompter Prompter,
	prefixFormat string,
	setter *protobuf.MessageSetter,
	ancestor []string,
	hasAncestorAndHasRepeatedField bool,
	color prompt.Color,
) *fieldInputter {
	return &fieldInputter{
		prompt:       prompter,
		setter:       setter,
		prefixFormat: prefixFormat,
		ancestor:     ancestor,
		color:        color,
		hasAncestorAndHasRepeatedField: hasAncestorAndHasRepeatedField,
	}
}

// Input will call itself for nested messages
//
// e.g.
// message Foo {
// 	Bar bar = 1;
//	string baz = 2;
// }
//
// Input is called two times
// one for Foo's primitive fields
// another one for bar's primitive fields
func (i *fieldInputter) Input(fields []entity.Field) (proto.Message, error) {
	if err := i.prompt.SetPrefixColor(i.color); err != nil {
		return nil, err
	}

	for _, field := range fields {
		if field.IsRepeated() {
			if err := i.inputRepeatedField(field); err != nil {
				return nil, err
			}
		} else {
			// if oneof, choose one from selection
			if oneof, ok := field.(entity.OneOfField); ok {
				var err error
				field, err = i.chooseOneof(oneof)
				if err != nil {
					return nil, err
				}
			}

			if err := i.inputField(field); err != nil {
				return nil, err
			}
		}
	}

	return i.setter.Done(), nil
}

func (i *fieldInputter) chooseOneof(oneof entity.OneOfField) (entity.Field, error) {
	options := make([]string, 0, len(oneof.Choices()))
	fieldOf := map[string]entity.Field{}
	for _, choice := range oneof.Choices() {
		options = append(options, choice.FieldName())
		fieldOf[choice.FieldName()] = choice
	}

	choice, err := i.prompt.Select(oneof.FieldName(), options)
	if err != nil {
		return nil, err
	}

	f, ok := fieldOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownOneofFieldName, choice)
	}

	return f, nil
}

func (i *fieldInputter) chooseEnum(enum entity.EnumField) (entity.EnumValue, error) {
	options := make([]string, 0, len(enum.Values()))
	valOf := map[string]entity.EnumValue{}
	for _, v := range enum.Values() {
		options = append(options, v.Name())
		valOf[v.Name()] = v
	}

	choice, err := i.prompt.Select(enum.Name(), options)
	if err != nil {
		return nil, err
	}

	c, ok := valOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownEnumName, choice)
	}

	return c, nil
}

func (i *fieldInputter) inputField(field entity.Field) error {
	switch f := field.(type) {
	case entity.EnumField:
		v, err := i.chooseEnum(f)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, v.Number()); err != nil {
			return err
		}
	case entity.MessageField:
		setter := protobuf.NewMessageSetter(f)
		fields := f.Fields()

		msg, err := newFieldInputter(i.prompt, i.prefixFormat, setter, append(i.ancestor, f.FieldName()), i.hasAncestorAndHasRepeatedField || f.IsRepeated(), i.color).Input(fields)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, msg); err != nil {
			return err
		}
		// increment prompt color to next one
		i.color = (i.color + 1) % 16
	case entity.PrimitiveField:
		i.prompt.SetPrefix(i.makePrefix(field))
		v, err := i.inputPrimitiveField(f)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, v); err != nil {
			return err
		}
	default:
		panic("unknown type: " + field.PBType())
	}
	return nil
}

func (i *fieldInputter) inputRepeatedField(f entity.Field) error {
	old := i.prompt
	defer func() {
		i.prompt = old
	}()
	// if repeated fields, create new prompt.
	// and the prompt will be terminate with ctrl+d.
	for {
		i.prompt = NewRealPrompter(nil, nil)
		i.prompt.SetPrefixColor(i.color)

		if err := i.inputField(f); err == EORF || err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		i.color = (i.color + 1) % 16
	}
	i.enteredEmptyInput = false
	return nil
}

func (i *fieldInputter) inputPrimitiveField(f entity.PrimitiveField) (interface{}, error) {
	l, err := i.prompt.Input()
	if err != nil {
		return "", err
	}

	part, err := shellstring.Parse(l)
	if err != nil {
		return "", err
	}

	if len(part) > 1 {
		return nil, errors.New("invalid input string")
	}

	var in string
	if len(part) != 0 {
		in = part[0]
	}

	if in == "" {
		// if f is repeated or
		// ancestor has repeated field
		if f.IsRepeated() || i.hasAncestorAndHasRepeatedField {
			if i.enteredEmptyInput {
				return nil, EORF
			}
			i.enteredEmptyInput = true
			// ignore the input
			return i.inputPrimitiveField(f)
		}
	} else {
		i.enteredEmptyInput = false
	}

	return protobuf.ConvertValue(in, f)
}

// makePrefix makes prefix for field f.
func (i *fieldInputter) makePrefix(f entity.PrimitiveField) string {
	return makePrefix(i.prefixFormat, f, i.ancestor, i.hasAncestorAndHasRepeatedField)
}

const (
	repeatedStr       = "<repeated> "
	ancestorDelimiter = "::"
)

func makePrefix(s string, f entity.PrimitiveField, ancestor []string, ancestorHasRepeated bool) string {
	joinedAncestor := strings.Join(ancestor, ancestorDelimiter)
	if joinedAncestor != "" {
		joinedAncestor += ancestorDelimiter
	}

	s = strings.Replace(s, "{ancestor}", joinedAncestor, -1)
	s = strings.Replace(s, "{name}", f.FieldName(), -1)
	s = strings.Replace(s, "{type}", f.PBType(), -1)

	if f.IsRepeated() || ancestorHasRepeated {
		return repeatedStr + s
	}
	return s
}
