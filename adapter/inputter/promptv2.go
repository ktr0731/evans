package inputter

import (
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/color"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
)

var (
	EORF = errors.New("end of repeated field")
)

var initialPromptInputterState = promptInputterState{
	selectedOneOf: make(map[string]interface{}),
	color:         color.DefaultColor(),
}

type promptInputterState struct {
	// Key: FullQualifiedName, Val: nil
	selectedOneOf map[string]interface{}

	ancestor []string
	color    color.Color
	// enteredEmptyInput is used to terminate repeated field inputting.
	// If input is empty and enteredEmptyInput is true, exit repeated input prompt.
	enteredEmptyInput bool
	// hasDirectCycledParent is true if the direct parent to the current field is a cycled message.
	hasDirectCycledParent bool
	// The field has parent fields and one or more fields is/are repeated.
	hasAncestorAndHasRepeatedField bool
}

// PromptInputter2 is an implementation of inputting method.
// it has common logic to input fields interactively.
// in normal, go-prompt is used as prompt.
type PromptInputter2 struct {
	prompt       prompt.Prompt
	prefixFormat string
	state        promptInputterState
}

func NewPrompt2(prefixFormat string) *PromptInputter2 {
	return &PromptInputter2{
		prompt:       prompt.New(nil, nil),
		prefixFormat: prefixFormat,
		state:        initialPromptInputterState,
	}
}

// Input receives a Protocol Buffers message descriptor and input each fields
// by using a prompt interactively.
// Returned proto.Message is a message that is set each field value inputted by a prompt.
//
// Note that Input resets the previous state when it is called again.
//
// Input is an implementation of port.Inputter.
func (i *PromptInputter2) Input(req *desc.MessageDescriptor) (proto.Message, error) {
	i.state = initialPromptInputterState
	m, err := i.inputMessage(req)
	if err != nil {
		if e, ok := errors.Cause(err).(*protobuf.ConversionError); ok {
			return nil, errors.Errorf("input '%s' is invalid in type %s", e.Val, e.ExpectedType)
		} else {
			return nil, err
		}
	}
	return m, nil
}

// inputMessage might call itself for nested messages.
//
// e.g.
//  message Foo {
//    Bar bar = 1;
//    string baz = 2;
//  }
//
// inputMessage is called two times.
// One for Foo's primitive fields, and another one for bar's primitive fields.
func (i *PromptInputter2) inputMessage(msg *desc.MessageDescriptor) (proto.Message, error) {
	if err := i.prompt.SetPrefixColor(i.state.color); err != nil {
		return nil, err
	}

	dmsg := dynamic.NewMessage(msg)

	for _, field := range msg.GetFields() {
		err := i.inputField(dmsg, field)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set inputted values to message '%s'", msg.GetFullyQualifiedName())
		}
	}

	return dmsg, nil
}

func (i *PromptInputter2) inputField(dmsg *dynamic.Message, f *desc.FieldDescriptor) error {
	if f.IsRepeated() {
		return i.inputRepeatedField(dmsg, f)
	}

	if isOneOfField(f) {
		if i.isSelectedOneOf(f) {
			return nil
		}
		var err error
		f, err = i.selectOneOf(f)
		if err != nil {
			return err
		}
		i.state.selectedOneOf[f.GetOneOf().GetFullyQualifiedName()] = nil
	}

	switch {
	// case entity.EnumField:
	// 	v, err := i.selectEnum(f)
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if err := i.setter.SetField(f, v.Number()); err != nil {
	// 		return err
	// 	}
	case f.GetMessageType() != nil:
		// if f.IsCycled() {
		// 	prefix := strings.Join(i.ancestor, ancestorDelimiter)
		// 	if prefix != "" {
		// 		prefix += ancestorDelimiter
		// 	}
		// 	prefix += f.FieldName()
		//
		// 	choice, err := i.prompt.Select(
		// 		fmt.Sprintf("circulated field was found. dig down or finish?\nfield: %s (%s)", prefix, f.FQRN()),
		// 		[]string{"dig down", "finish"},
		// 	)
		// 	if err != nil {
		// 		return err
		// 	}
		// 	if choice == "finish" {
		// 		if f.IsRepeated() {
		// 			return EORF
		// 		}
		// 		return nil
		// 	}
		// 	// TODO: coloring
		// 	// i.color.next()
		// }
		msg, err := i.inputMessage(f.GetMessageType())
		// msg, err := newFieldInputter(
		// 	i.prompt,
		// 	i.prefixFormat,
		// 	setter,
		// 	append(i.ancestor, f.FieldName()),
		// 	i.hasAncestorAndHasRepeatedField || f.IsRepeated(),
		// 	f.IsCycled(),
		// 	i.color,
		// ).Input(fields)
		if err != nil {
			return err
		}
		if err := dmsg.TrySetField(f, msg); err != nil {
			return errors.Wrap(err, "failed to set an inputted message to a field")
		}
		i.state.color.Next()
	default: // Normal fields.
		i.prompt.SetPrefix(i.makePrefix(f))
		v, err := i.inputPrimitiveField(f)
		if err != nil {
			return err
		}

		if err := dmsg.TrySetField(f, v); err != nil {
			return errors.Wrapf(err, "failed to set inputted value to field '%s'", f.GetName())
		}
	}

	return nil
}

func (i *PromptInputter2) selectOneOf(f *desc.FieldDescriptor) (*desc.FieldDescriptor, error) {
	oneof := f.GetOneOf()

	options := make([]string, len(oneof.GetChoices()))
	fieldOf := map[string]*desc.FieldDescriptor{}
	for i, choice := range oneof.GetChoices() {
		options[i] = choice.GetName()
		fieldOf[choice.GetName()] = choice
	}

	choice, err := i.prompt.Select(oneof.GetName(), options)
	if err != nil {
		return nil, err
	}

	return fieldOf[choice], nil
}

func (i *PromptInputter2) selectEnum(enum entity.EnumField) (entity.EnumValue, error) {
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
		return nil, errors.Errorf("unknown enum '%s'", choice)
	}

	return c, nil
}

// func (i *PromptInputter2) inputField(field entity.Field) error {
// 	switch f := field.(type) {
// 	case entity.EnumField:
// 		v, err := i.selectEnum(f)
// 		if err != nil {
// 			return err
// 		}
// 		if err := i.setter.SetField(f, v.Number()); err != nil {
// 			return err
// 		}
// 	case entity.MessageField:
// 		if f.IsCycled() {
// 			prefix := strings.Join(i.ancestor, ancestorDelimiter)
// 			if prefix != "" {
// 				prefix += ancestorDelimiter
// 			}
// 			prefix += f.FieldName()
//
// 			choice, err := i.prompt.Select(
// 				fmt.Sprintf("circulated field was found. dig down or finish?\nfield: %s (%s)", prefix, f.FQRN()),
// 				[]string{"dig down", "finish"},
// 			)
// 			if err != nil {
// 				return err
// 			}
// 			if choice == "finish" {
// 				if f.IsRepeated() {
// 					return EORF
// 				}
// 				return nil
// 			}
// 			// TODO: coloring
// 			// i.color.next()
// 		}
// 		setter := protobuf.NewMessageSetter(f)
// 		fields := f.Fields()
//
// 		msg, err := newFieldInputter(
// 			i.prompt,
// 			i.prefixFormat,
// 			setter,
// 			append(i.ancestor, f.FieldName()),
// 			i.hasAncestorAndHasRepeatedField || f.IsRepeated(),
// 			f.IsCycled(),
// 			i.color,
// 		).Input(fields)
// 		if err != nil {
// 			return err
// 		}
// 		if err := i.setter.SetField(f, msg); err != nil {
// 			return err
// 		}
// 		i.color.Next()
// 	case entity.PrimitiveField:
// 		i.prompt.SetPrefix(i.makePrefix(field))
// 		v, err := i.inputPrimitiveField(f)
// 		if err != nil {
// 			return err
// 		}
// 		if err := i.setter.SetField(f, v); err != nil {
// 			return err
// 		}
// 	default:
// 		panic("unknown type: " + field.PBType())
// 	}
// 	return nil
// }

func (i *PromptInputter2) inputRepeatedField(dmsg *dynamic.Message, f *desc.FieldDescriptor) error {
	old := i.prompt
	defer func() {
		i.prompt = old
	}()
	// If repeated fields, create new prompt.
	// The prompt will be terminate with ctrl+d.
	for {
		i.prompt = prompt.New(nil, nil)
		if err := i.prompt.SetPrefixColor(i.state.color); err != nil {
			return err
		}

		if err := i.inputField(dmsg, f); err == EORF || err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}

		i.state.color.Next()
	}
}

func (i *PromptInputter2) inputPrimitiveField(f *desc.FieldDescriptor) (interface{}, error) {
	in, err := i.prompt.Input()
	if err != nil {
		return "", errors.Wrap(err, "failed to read user input")
	}

	// Empty input. See enteredEmptyInput field comments for the behavior
	// when empty input is entered.
	if in == "" {
		// Check whether f.enteredEmptyInput is true or false when this field satisfies one of following conditions.
		//   - f is a repeated field
		//   - One or more ancestor has/have repeated field and
		//     the direct parent is not a cycled field.
		if f.IsRepeated() || (i.state.hasAncestorAndHasRepeatedField && !i.state.hasDirectCycledParent) {
			if i.state.enteredEmptyInput {
				return nil, EORF
			}
			i.state.enteredEmptyInput = true
			// ignore the input
			return i.inputPrimitiveField(f)
		}
	} else {
		i.state.enteredEmptyInput = false
	}

	return protobuf.ConvertValue2(in, f)
}

func (i *PromptInputter2) isSelectedOneOf(f *desc.FieldDescriptor) bool {
	_, ok := i.state.selectedOneOf[f.GetOneOf().GetFullyQualifiedName()]
	return ok
}

// makePrefix makes prefix for field f.
func (i *PromptInputter2) makePrefix(f *desc.FieldDescriptor) string {
	return makePrefix2(i.prefixFormat, f, i.state.ancestor, i.state.hasAncestorAndHasRepeatedField)
}

const (
	repeatedStr       = "<repeated> "
	ancestorDelimiter = "::"
)

func makePrefix2(s string, f *desc.FieldDescriptor, ancestor []string, ancestorHasRepeated bool) string {
	joinedAncestor := strings.Join(ancestor, ancestorDelimiter)
	if joinedAncestor != "" {
		joinedAncestor += ancestorDelimiter
	}

	s = strings.Replace(s, "{ancestor}", joinedAncestor, -1)
	s = strings.Replace(s, "{name}", f.GetName(), -1)
	s = strings.Replace(s, "{type}", f.GetType().String(), -1)

	if f.IsRepeated() || ancestorHasRepeated {
		return repeatedStr + s
	}
	return s
}

func isOneOfField(f *desc.FieldDescriptor) bool {
	return f.GetOneOf() != nil
}
