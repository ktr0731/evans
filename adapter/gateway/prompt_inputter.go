package gateway

import (
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
	Input() string
	Select(msg string, opts []string) (string, error)
	SetPrefix(prefix string)
	SetPrefixColor(color prompt.Color) error
}

type RealPrompter struct {
	fieldPrompter *prompt.Prompt
	currentPrefix string
}

func newRealPrompter() *RealPrompter {
	executor := func(in string) {
		return
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return nil
	}
	p := &RealPrompter{}
	p.fieldPrompter = prompt.New(executor, completer, prompt.OptionLivePrefix(p.livePrefix))
	return p
}

func (p *RealPrompter) Input() string {
	return p.fieldPrompter.Input()
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

// mixin go-prompt
var NewPrompt = func(config *config.Config, env entity.Environment) *Prompt {
	return newPrompt(newRealPrompter(), config, env)
}

// Prompt has common logic to input fields interactively.
// prompt is an implementation of inputting method.
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
			for {
				if err := i.inputField(field); err == EORF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			i.enteredEmptyInput = false
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

func (i *fieldInputter) inputPrimitiveField(f entity.PrimitiveField) (interface{}, error) {
	l := i.prompt.Input()
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
