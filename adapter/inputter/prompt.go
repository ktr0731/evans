package inputter

import (
	"fmt"
	"io"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/adapter/prompt"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/color"
	"github.com/pkg/errors"
)

var initialPromptInputterState = promptInputterState{
	selectedOneOf:      make(map[string]interface{}),
	circulatedMessages: make(map[string][]string),
	color:              color.DefaultColor(),
}

// promptInputterState holds states for inputting a message.
type promptInputterState struct {
	// Key: fully-qualified message name, Val: nil
	selectedOneOf map[string]interface{}
	// circulatedMessages holds a fully-qualified message name,
	// and the val represents whether the message circulates or not.
	// If a message circulates, val holds a slice of names from the key message
	// until the last message.
	// If a message doesn't circulate, the val is nil.
	//
	// A key is assigned at calling a RPC that requires the message.
	circulatedMessages map[string][]string

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

// clone copies itself. Copied map fields refer to another position from the original.
func (s promptInputterState) clone() promptInputterState {
	new := s
	newSelectedOneOf := make(map[string]interface{})
	for k, v := range s.selectedOneOf {
		newSelectedOneOf[k] = v
	}
	new.selectedOneOf = newSelectedOneOf
	newCirculatedMessages := make(map[string][]string)
	for k, v := range s.circulatedMessages {
		newCirculatedMessages[k] = v
	}
	new.circulatedMessages = newCirculatedMessages
	return new
}

// PromptInputter is an implementation of inputting method.
// it has common logic to input fields interactively.
// in normal, go-prompt is used as prompt.
type PromptInputter struct {
	prompt       prompt.Prompt
	prefixFormat string
	state        promptInputterState
}

func NewPrompt(prefixFormat string) *PromptInputter {
	return newPrompt(prompt.New(nil, nil), prefixFormat)
}

// For testing, the real constructor is separated from NewPrompt.
func newPrompt(prompt prompt.Prompt, prefixFormat string) *PromptInputter {
	return &PromptInputter{
		prompt:       prompt,
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
func (i *PromptInputter) Input(req *desc.MessageDescriptor) (proto.Message, error) {
	i.state = initialPromptInputterState.clone()
	m, err := i.inputMessage(req)
	if err != nil {
		// If io.EOF is returned, it means CTRL+d is entered.
		// In this case, Input skips rest fields and finishes normally.
		if err == io.EOF {
			return m, nil
		}
		if e, ok := errors.Cause(err).(*protobuf.ConversionError); ok {
			return nil, errors.Errorf("input '%s' is invalid in type %s", e.Val, e.ExpectedType)
		}
		return nil, err
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
//
// inputMessage returns following errors:
//
//   - io.EOF:
//       CTRL+d entered. Never return in the case of repeated message.
//       inputMessage also returns the first return value what is
//       the message partially inputted.
func (i *PromptInputter) inputMessage(msg *desc.MessageDescriptor) (proto.Message, error) {
	if err := i.prompt.SetPrefixColor(i.state.color); err != nil {
		return nil, err
	}

	dmsg := dynamic.NewMessage(msg)

	// i.state is message-scoped, so we reset i.state after inputting msg.
	currentState := i.state.clone()
	defer func() {
		i.state = currentState
	}()

	for _, field := range msg.GetFields() {
		err := i.inputField(dmsg, field, false)
		if errors.Cause(err) == io.EOF {
			return dmsg, io.EOF
		}
		if err != nil {
			return nil, errors.Wrapf(err, "failed to set inputted values to message '%s'", msg.GetFullyQualifiedName())
		}
	}

	return dmsg, nil
}

// inputField tries to set a inputted value to a field of the passed message dmsg.
// An argument partOfRepeatedField means inputField is called from inputRepeatedField.
//
// inputField returns following errors:
//   - io.EOF: CTRL+d is entered.
func (i *PromptInputter) inputField(dmsg *dynamic.Message, f *desc.FieldDescriptor, partOfRepeatedField bool) error {
	// If a repeated field is found, call inputRepeatedField instead.
	if !partOfRepeatedField && f.IsRepeated() {
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

	old := i.state.hasAncestorAndHasRepeatedField
	i.state.hasAncestorAndHasRepeatedField = i.state.hasAncestorAndHasRepeatedField || f.IsRepeated()
	defer func() {
		i.state.hasAncestorAndHasRepeatedField = old
	}()

	switch {
	case f.GetEnumType() != nil:
		v, err := i.selectEnum(f)
		if err != nil {
			return err
		}
		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(f, v.GetNumber()); err != nil {
				return err
			}
		} else {
			if err := dmsg.TrySetField(f, v.GetNumber()); err != nil {
				return err
			}
		}
	case f.GetMessageType() != nil:
		if i.isCirculatedField(f) {
			prefix := strings.Join(i.state.ancestor, ancestorDelimiter)
			if prefix != "" {
				prefix += ancestorDelimiter
			}
			prefix += f.GetName()

			choice, err := i.prompt.Select(
				fmt.Sprintf("circulated field was found. dig down or finish?\nfield: %s (%s)", prefix, strings.Join(i.state.circulatedMessages[f.GetMessageType().GetFullyQualifiedName()], ">")),
				[]string{"dig down", "finish"},
			)
			if err != nil {
				return err
			}
			if choice == "finish" {
				if f.IsRepeated() {
					return io.EOF
				}
				return nil
			}
		}

		ancestorLen := len(i.state.ancestor)
		i.state.ancestor = append(i.state.ancestor, f.GetName())

		msg, err := i.inputMessage(f.GetMessageType())
		// If io.EOF is returned, msg isn't nil (see inputMessage comments).
		if err != nil && err != io.EOF {
			return err
		}

		// Don't set empty messages.
		emptyMsg := len(msg.String()) == 0
		if emptyMsg && err == io.EOF {
			return io.EOF
		}

		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(f, msg); err != nil {
				return errors.Wrap(err, "failed to add an inputted message to a repeated field")
			}
		} else {
			if err := dmsg.TrySetField(f, msg); err != nil {
				return errors.Wrap(err, "failed to set an inputted message to a field")
			}
		}

		// If err is io.EOF, propagate it to the caller after TrySetField/TryAddRepeatedField.
		if err == io.EOF {
			return err
		}

		// Discard appended ancestors after calling above inputMessage.
		i.state.ancestor = i.state.ancestor[:ancestorLen]
		i.state.color.Next()
	default: // Normal fields.
		i.prompt.SetPrefix(i.makePrefix(f))
		v, err := i.inputPrimitiveField(f)
		if err != nil {
			return err
		}

		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(f, v); err != nil {
				return errors.Wrapf(err, "failed to add inputted value to repeated field '%s'", f.GetName())
			}
		} else {
			if err := dmsg.TrySetField(f, v); err != nil {
				return errors.Wrapf(err, "failed to set inputted value to field '%s'", f.GetName())
			}
		}
	}

	return nil
}

func (i *PromptInputter) selectOneOf(f *desc.FieldDescriptor) (*desc.FieldDescriptor, error) {
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

	desc, ok := fieldOf[choice]
	if !ok {
		return nil, errors.Errorf("invalid choice '%s' selected", choice)
	}

	return desc, nil
}

func (i *PromptInputter) selectEnum(enum *desc.FieldDescriptor) (*desc.EnumValueDescriptor, error) {
	values := enum.GetEnumType().GetValues()
	options := make([]string, 0, len(values))
	valOf := map[string]*desc.EnumValueDescriptor{}
	for _, v := range values {
		options = append(options, v.GetName())
		valOf[v.GetName()] = v
	}

	choice, err := i.prompt.Select(enum.GetName(), options)
	if err != nil {
		return nil, err
	}

	c, ok := valOf[choice]
	if !ok {
		return nil, errors.Errorf("unknown enum '%s'", choice)
	}

	return c, nil
}

func (i *PromptInputter) inputRepeatedField(dmsg *dynamic.Message, f *desc.FieldDescriptor) error {
	old := i.prompt
	defer func() {
		i.prompt = old
	}()
	i.prompt = prompt.New(nil, nil)
	// If repeated fields, create new prompt.
	// The prompt will be terminate with ctrl+d.
	for {
		if err := i.prompt.SetPrefixColor(i.state.color); err != nil {
			return err
		}

		err := i.inputField(dmsg, f, true)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		i.state.color.Next()
	}
}

// inputPrimitiveField reads an input and converts it to a Go type.
// If CTRL+d is entered, inputPrimitiveField returns io.EOF.
func (i *PromptInputter) inputPrimitiveField(f *desc.FieldDescriptor) (interface{}, error) {
	in, err := i.prompt.Input()
	if err == io.EOF {
		return "", io.EOF
	}
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
				return nil, io.EOF
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

func (i *PromptInputter) isSelectedOneOf(f *desc.FieldDescriptor) bool {
	_, ok := i.state.selectedOneOf[f.GetOneOf().GetFullyQualifiedName()]
	return ok
}

// isCirculatedField checks whether the passed message m is a circulated message.
// TODO: a message is possible to have one or more circulated fields.
// Field f must be a message field.
func (i *PromptInputter) isCirculatedField(f *desc.FieldDescriptor) bool {
	var appearedFields []*desc.FieldDescriptor
	appeared := make(map[string]interface{})

	// TODO: checkCirculatedField という名前はよくない
	var checkCirculatedField func(f *desc.FieldDescriptor)
	checkCirculatedField = func(f *desc.FieldDescriptor) {
		appeared[f.GetMessageType().GetFullyQualifiedName()] = nil
		appearedFields = append(appearedFields, f)
		copiedAppeared := make(map[string]interface{})
		for k, v := range appeared {
			copiedAppeared[k] = v
		}
		copiedAppearedFields := make([]*desc.FieldDescriptor, len(appearedFields))
		copy(copiedAppearedFields, appearedFields)

		defer func() {
			appeared = copiedAppeared
			appearedFields = copiedAppearedFields
		}()

		for _, field := range f.GetMessageType().GetFields() {
			msg := field.GetMessageType()
			if msg == nil {
				continue
			}
			msgName := msg.GetFullyQualifiedName()
			// もし message がすでに現れていたら循環している
			// appearedField の先頭から舐めて重複したフィールドを見つけ、
			// それ以降を循環している部分とする
			if _, found := appeared[msgName]; found {
				for idx := 0; idx < len(appearedFields); idx++ {
					if appearedFields[idx].GetMessageType().GetFullyQualifiedName() != msgName {
						continue
					}
					appearedMsgs := make([]string, len(appearedFields[idx:]))
					for i, f := range appearedFields[idx:] {
						appearedMsgs[i] = f.GetMessageType().GetFullyQualifiedName()
					}
					i.state.circulatedMessages[appearedFields[idx].GetFullyQualifiedName()] = appearedMsgs
				}
				return
			}
			checkCirculatedField(field)
		}
	}

	checkCirculatedField(f)
	return len(i.state.circulatedMessages[f.GetFullyQualifiedName()]) != 0
}

// makePrefix makes prefix for field f.
func (i *PromptInputter) makePrefix(f *desc.FieldDescriptor) string {
	return makePrefix(i.prefixFormat, f, i.state.ancestor, i.state.hasAncestorAndHasRepeatedField)
}

const (
	repeatedStr       = "<repeated> "
	ancestorDelimiter = "::"
)

func makePrefix(s string, f *desc.FieldDescriptor, ancestor []string, ancestorHasRepeated bool) string {
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
