package proto

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
	"github.com/pkg/errors"
)

// InteractiveFiller is an implementation of fill.InteractiveFiller.
// It let you input request fields interactively.
type InteractiveFiller struct {
	prompt       prompt.Prompt
	prefixFormat string
	state        promptInputterState

	digManually   bool
	bytesFromFile bool
}

// NewInteractiveFiller instantiates a new filler that fills each field interactively.
func NewInteractiveFiller(prompt prompt.Prompt, prefixFormat string) *InteractiveFiller {
	return &InteractiveFiller{
		prompt:       prompt,
		prefixFormat: prefixFormat,
	}
}

// Fill receives v that is an instance of *dynamic.Message.
// Fill let you input each field interactively by using a prompt. v will be set field values inputted by a prompt.
//
// Note that Fill resets the previous state when it is called again.
func (f *InteractiveFiller) Fill(v interface{}, opts fill.InteractiveFillerOpts) error {
	f.digManually = opts.DigManually
	f.bytesFromFile = opts.BytesFromFile

	msg, ok := v.(*dynamic.Message)
	if !ok {
		return fill.ErrCodecMismatch
	}

	f.state = initialPromptInputterState.clone()
	err := f.inputMessage(msg)
	// If io.EOF is returned, it means CTRL+d is entered.
	// In this case, Input skips rest fields and finishes normally.
	if errors.Is(err, io.EOF) {
		return io.EOF
	}
	if errors.Is(err, prompt.ErrAbort) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

// inputMessage might call itself for nested messages.
//
// e.g.
//  message Foo {
//    Bar bar = 1;
//    string baz = 2;
//  }
//
// inputMessage is called two times. One for Foo's primitive fields, and another one for bar's primitive fields.
//
// inputMessage returns following errors:
//
//   - io.EOF:
//       CTRL+d entered. Never return in the case of repeated message.
//       inputMessage also returns the first return value what is the message partially inputted.
//
func (f *InteractiveFiller) inputMessage(dmsg *dynamic.Message) error {
	f.prompt.SetPrefixColor(f.state.color)

	// f.state is message-scoped, so we reset f.state after inputting msg.
	currentState := f.state.clone()
	defer func() {
		f.state = currentState
	}()

	for _, field := range dmsg.GetMessageDescriptor().GetFields() {
		err := f.inputField(dmsg, field, false)
		if errors.Is(err, io.EOF) {
			return io.EOF
		}
		if err != nil {
			return errors.Wrapf(err, "failed to set inputted values to message '%s'", dmsg.GetMessageDescriptor().GetFullyQualifiedName())
		}
	}

	return nil
}

// inputField tries to set a inputted value to a field of the passed message dmsg.
// An argument partOfRepeatedField means inputField is called from inputRepeatedField.
//
// inputField returns following errors:
//   - io.EOF: CTRL+d is entered.
func (f *InteractiveFiller) inputField(dmsg *dynamic.Message, field *desc.FieldDescriptor, partOfRepeatedField bool) error {
	// If a repeated field is found, call inputRepeatedField instead.
	if !partOfRepeatedField && field.IsRepeated() {
		return f.inputRepeatedField(dmsg, field)
	}

	if isOneOfField(field) {
		if f.isSelectedOneOf(field) {
			return nil
		}
		var err error
		field, err = f.selectOneOf(field)
		if err != nil {
			return err
		}
		f.state.selectedOneOf[field.GetOneOf().GetFullyQualifiedName()] = nil
	}

	old := f.state.hasAncestorAndHasRepeatedField
	f.state.hasAncestorAndHasRepeatedField = f.state.hasAncestorAndHasRepeatedField || field.IsRepeated()
	defer func() {
		f.state.hasAncestorAndHasRepeatedField = old
	}()

	switch {
	case field.GetEnumType() != nil:
		v, err := f.selectEnum(field)
		if err != nil {
			return err
		}
		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(field, v.GetNumber()); err != nil {
				return err
			}
		} else {
			if err := dmsg.TrySetField(field, v.GetNumber()); err != nil {
				return err
			}
		}
	case field.GetMessageType() != nil:
		if f.isCirculatedField(field) {
			prefix := strings.Join(f.state.ancestor, ancestorDelimiter)
			if prefix != "" {
				prefix += ancestorDelimiter
			}
			prefix += field.GetName()

			choice, err := f.prompt.Select(
				fmt.Sprintf("circulated field was found. dig down or finish? field: %s (%s)", prefix, strings.Join(f.state.circulatedMessages[field.GetMessageType().GetFullyQualifiedName()], ">")),
				[]string{"dig down", "finish"},
			)
			if err != nil {
				return err
			}
			if choice == "finish" {
				if field.IsRepeated() {
					return io.EOF
				}
				return nil
			}
		}

		ancestorLen := len(f.state.ancestor)
		f.state.ancestor = append(f.state.ancestor, field.GetName())

		if f.digManually {
			choice, err := f.prompt.Select(
				fmt.Sprintf(
					"dig down? field: %s, message: %s",
					field.GetFullyQualifiedName(),
					dmsg.GetMessageDescriptor().GetFullyQualifiedName(),
				),
				[]string{"dig down", "skip"},
			)
			if err != nil {
				return err
			}
			if choice == "skip" {
				return nil
			}
		}

		msg := dynamic.NewMessage(field.GetMessageType())
		err := f.inputMessage(msg)
		// If io.EOF is returned, msg isn't nil (see inputMessage comments).
		if err != nil && !errors.Is(err, prompt.ErrAbort) {
			return err
		}
		if errors.Is(err, io.EOF) {
			return io.EOF
		}

		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(field, msg); err != nil {
				return errors.Wrap(err, "failed to add an inputted message to a repeated field")
			}
		} else {
			if err := dmsg.TrySetField(field, msg); err != nil {
				return errors.Wrap(err, "failed to set an inputted message to a field")
			}
		}

		if partOfRepeatedField && errors.Is(err, prompt.ErrAbort) {
			return prompt.ErrAbort
		}

		// Discard appended ancestors after calling above inputMessage.
		f.state.ancestor = f.state.ancestor[:ancestorLen]
		f.state.color.Next()
	default: // Normal fields.
		f.prompt.SetPrefix(f.makePrefix(field))
		v, err := f.inputPrimitiveField(field.GetType())
		if err != nil {
			return err
		}

		if partOfRepeatedField {
			if err := dmsg.TryAddRepeatedField(field, v); err != nil {
				return errors.Wrapf(err, "failed to add inputted value to repeated field '%s'", field.GetName())
			}
		} else {
			if err := dmsg.TrySetField(field, v); err != nil {
				return errors.Wrapf(err, "failed to set inputted value to field '%s'", field.GetName())
			}
		}
	}

	return nil
}

func (f *InteractiveFiller) selectOneOf(field *desc.FieldDescriptor) (*desc.FieldDescriptor, error) {
	oneof := field.GetOneOf()

	options := make([]string, len(oneof.GetChoices()))
	fieldOf := map[string]*desc.FieldDescriptor{}
	for i, choice := range oneof.GetChoices() {
		options[i] = choice.GetName()
		fieldOf[choice.GetName()] = choice
	}

	choice, err := f.prompt.Select(oneof.GetName(), options)
	if err != nil {
		return nil, err
	}

	desc, ok := fieldOf[choice]
	if !ok {
		return nil, errors.Errorf("invalid choice '%s' selected", choice)
	}

	return desc, nil
}

func (f *InteractiveFiller) selectEnum(enum *desc.FieldDescriptor) (*desc.EnumValueDescriptor, error) {
	values := enum.GetEnumType().GetValues()
	options := make([]string, 0, len(values))
	valOf := map[string]*desc.EnumValueDescriptor{}
	for _, v := range values {
		options = append(options, v.GetName())
		valOf[v.GetName()] = v
	}

	choice, err := f.prompt.Select(enum.GetName(), options)
	if err != nil {
		return nil, err
	}

	c, ok := valOf[choice]
	if !ok {
		return nil, errors.Errorf("unknown enum '%s'", choice)
	}

	return c, nil
}

func (f *InteractiveFiller) inputRepeatedField(dmsg *dynamic.Message, field *desc.FieldDescriptor) error {
	old := f.prompt
	defer func() {
		f.prompt = old
	}()

	// If repeated fields, create new prompt. The prompt will be terminate with ctrl+d.
	for {
		f.prompt.SetPrefixColor(f.state.color)

		err := f.inputField(dmsg, field, true)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		f.state.color.Next()
	}
}

// inputPrimitiveField reads an input and converts it to a Go type.
// If CTRL+d is entered, inputPrimitiveField returns io.EOF.
func (f *InteractiveFiller) inputPrimitiveField(fieldType descriptor.FieldDescriptorProto_Type) (interface{}, error) {
	in, err := f.prompt.Input()
	if errors.Is(err, io.EOF) {
		return "", io.EOF
	}
	if err != nil {
		return "", errors.Wrap(err, "failed to read user input")
	}

	v, err := convertValue(in, fieldType)
	if err != nil {
		return nil, err
	}

	if fieldType == descriptor.FieldDescriptorProto_TYPE_BYTES {
		return f.processBytesInput(v)
	}

	return v, err
}

func (f *InteractiveFiller) processBytesInput(v interface{}) (interface{}, error) {
	if _, ok := v.([]byte); !ok {
		return nil, errors.New("value is not of type bytes")
	}

	if f.bytesFromFile {
		return readFileFromRelativePath(string(v.([]byte)))
	}

	return v, nil
}

func (f *InteractiveFiller) isSelectedOneOf(field *desc.FieldDescriptor) bool {
	_, ok := f.state.selectedOneOf[field.GetOneOf().GetFullyQualifiedName()]
	return ok
}

// isCirculatedField checks whether the passed message m is a circulated message. Field f must be a message field.
func (f *InteractiveFiller) isCirculatedField(field *desc.FieldDescriptor) bool {
	var appearedFields []*desc.FieldDescriptor
	appeared := make(map[string]interface{})

	var checkCirculatedField func(*desc.FieldDescriptor)
	checkCirculatedField = func(field *desc.FieldDescriptor) {
		appeared[field.GetMessageType().GetFullyQualifiedName()] = nil
		appearedFields = append(appearedFields, field)
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

		for _, field := range field.GetMessageType().GetFields() {
			msg := field.GetMessageType()
			if msg == nil {
				continue
			}
			msgName := msg.GetFullyQualifiedName()
			// If msgName is already appeared, the message is a circulated message.
			// Find the duplicate field from the top of appearedFields,
			// and let it be the part that circulates.
			if _, found := appeared[msgName]; found {
				for idx := 0; idx < len(appearedFields); idx++ {
					if appearedFields[idx].GetMessageType().GetFullyQualifiedName() != msgName {
						continue
					}
					appearedMsgs := make([]string, len(appearedFields[idx:]))
					for i, f := range appearedFields[idx:] {
						appearedMsgs[i] = f.GetMessageType().GetFullyQualifiedName()
					}
					f.state.circulatedMessages[appearedFields[idx].GetFullyQualifiedName()] = appearedMsgs
				}
				return
			}
			checkCirculatedField(field)
		}
	}

	checkCirculatedField(field)
	return len(f.state.circulatedMessages[field.GetFullyQualifiedName()]) != 0
}

// makePrefix makes prefix for field f.
func (f *InteractiveFiller) makePrefix(field *desc.FieldDescriptor) string {
	return makePrefix(f.prefixFormat, field, f.state.ancestor, f.state.hasAncestorAndHasRepeatedField)
}

var initialPromptInputterState = promptInputterState{
	selectedOneOf:      make(map[string]interface{}),
	circulatedMessages: make(map[string][]string),
	color:              prompt.ColorInitial,
}

// promptInputterState holds states for inputting a message.
type promptInputterState struct {
	// Key: fully-qualified message name, Val: nil
	selectedOneOf map[string]interface{}
	// circulatedMessages holds a fully-qualified message name,
	// and the val represents whether the message circulates or not.
	// If a message circulates, val holds a slice of names from the key message until the last message.
	// If a message doesn't circulate, the val is nil.
	//
	// A key is assigned at calling a RPC that requires the message.
	circulatedMessages map[string][]string

	ancestor []string
	color    prompt.Color
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

const (
	repeatedStr       = "<repeated> "
	ancestorDelimiter = "::"
)

func makePrefix(s string, field *desc.FieldDescriptor, ancestor []string, ancestorHasRepeated bool) string {
	joinedAncestor := strings.Join(ancestor, ancestorDelimiter)
	if joinedAncestor != "" {
		joinedAncestor += ancestorDelimiter
	}

	s = strings.ReplaceAll(s, "{ancestor}", joinedAncestor)
	s = strings.ReplaceAll(s, "{name}", field.GetName())
	s = strings.ReplaceAll(s, "{type}", field.GetType().String())

	if field.IsRepeated() || ancestorHasRepeated {
		return repeatedStr + s
	}
	return s
}

func isOneOfField(field *desc.FieldDescriptor) bool {
	return field.GetOneOf() != nil
}

func readFileFromRelativePath(path string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}
	return ioutil.ReadFile(path)
}
