package inputter

import (
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/color"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/entity/env"
	"github.com/pkg/errors"
)

var (
	ErrUnknownOneofFieldName = errors.New("unknown oneof field name")
	ErrUnknownEnumName       = errors.New("unknown enum name")
	EORF                     = errors.New("end of repeated field")
)

// PromptInputter is an implementation of inputting method.
// it has common logic to input fields interactively.
// in normal, go-prompt is used as prompt.
type PromptInputter struct {
	prompt       prompt.Prompt
	prefixFormat string
	env          env.Environment
}

func NewPrompt(prefixFormat string, env env.Environment) *PromptInputter {
	return newPromptInputter(prompt.New(nil, nil), prefixFormat, env)
}

func newPromptInputter(prompt prompt.Prompt, prefixFormat string, env env.Environment) *PromptInputter {
	return &PromptInputter{
		prompt:       prompt,
		prefixFormat: prefixFormat,
		env:          env,
	}
}

// Input is an implementation of port.Inputter
func (i *PromptInputter) Input(reqType entity.Message) (proto.Message, error) {
	setter := protobuf.NewMessageSetter(reqType)
	fields := reqType.Fields()

	// DarkGreen is the initial color
	return newFieldInputter(i.prompt, i.prefixFormat, setter, []string{}, false, color.DefaultColor()).Input(fields)
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
type fieldInputter struct {
	prompt prompt.Prompt
	setter *protobuf.MessageSetter

	prefixFormat string
	ancestor     []string

	color color.Color

	// enteredEmptyInput is used to terminate repeated field inputting
	// if input is empty and enteredEmptyInput is true, exit repeated input prompt
	enteredEmptyInput bool

	// the field has parent field (in other words, the field is child of a message field)
	hasAncestorAndHasRepeatedField bool
}

func newFieldInputter(
	prompter prompt.Prompt,
	prefixFormat string,
	setter *protobuf.MessageSetter,
	ancestor []string,
	hasAncestorAndHasRepeatedField bool,
	color color.Color,
) *fieldInputter {
	return &fieldInputter{
		prompt:                         prompter,
		setter:                         setter,
		prefixFormat:                   prefixFormat,
		ancestor:                       ancestor,
		color:                          color,
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
		if f.IsCycled() {
			prefix := strings.Join(i.ancestor, ancestorDelimiter)
			if prefix != "" {
				prefix += ancestorDelimiter
			}
			prefix += f.FieldName()

			choice, err := i.prompt.Select(
				fmt.Sprintf("circulated field was found. dig down or finish?\nfield: %s (%s)", prefix, f.FQRN()),
				[]string{"dig down", "finish"},
			)
			if err != nil {
				return err
			}
			if choice == "finish" {
				if f.IsRepeated() {
					return EORF
				}
				return nil
			}
			// TODO: coloring
			// i.color.next()
		}
		setter := protobuf.NewMessageSetter(f)
		fields := f.Fields()

		msg, err := newFieldInputter(
			i.prompt,
			i.prefixFormat,
			setter,
			append(i.ancestor, f.FieldName()),
			i.hasAncestorAndHasRepeatedField || f.IsRepeated(),
			i.color,
		).Input(fields)
		if err != nil {
			return err
		}
		if err := i.setter.SetField(f, msg); err != nil {
			return err
		}
		i.color.Next()
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
		i.prompt = prompt.New(nil, nil)
		if err := i.prompt.SetPrefixColor(i.color); err != nil {
			return err
		}

		if err := i.inputField(f); err == EORF || err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		i.color.Next()
	}
}

func (i *fieldInputter) inputPrimitiveField(f entity.PrimitiveField) (interface{}, error) {
	in, err := i.prompt.Input()
	if err != nil {
		return "", err
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
