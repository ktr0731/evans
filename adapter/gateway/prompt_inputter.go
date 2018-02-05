package gateway

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/config"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
)

var (
	ErrUnknownOneofFieldName = errors.New("unknown oneof field name")
	ErrUnknownEnumName       = errors.New("unknown enum name")
	EORF                     = errors.New("end of repeated field")
)

// for mocking
type prompter interface {
	Input() string
	Select(msg string, opts []string) (string, error)
	SetPrefix(prefix string) error
	SetPrefixColor(color prompt.Color) error
}

type RealPrompter struct {
	fieldPrompter *prompt.Prompt
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

func (p *RealPrompter) SetPrefix(prefix string) error {
	return prompt.OptionPrefix(prefix)(p.fieldPrompter)
}

func (p *RealPrompter) SetPrefixColor(color prompt.Color) error {
	return prompt.OptionPrefixTextColor(color)(p.fieldPrompter)
}

// mixin go-prompt
func NewPrompt(config *config.Config, env entity.Environment) *Prompt {
	executor := func(in string) {
		return
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return nil
	}
	return newPrompt(&RealPrompter{prompt.New(executor, completer)}, config, env)
}

type Prompt struct {
	prompt prompter
	config *config.Config
	env    entity.Environment
}

func newPrompt(prompt prompter, config *config.Config, env entity.Environment) *Prompt {
	return &Prompt{
		prompt: prompt,
		config: config,
		env:    env,
	}
}

func (i *Prompt) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	req := dynamic.NewMessage(reqType)
	fields := reqType.GetFields()

	// DarkGreen is the initial color
	return newFieldInputter(i.prompt, i.config.Env.InputPromptFormat, req, reqType, true, prompt.DarkGreen).Input(fields)
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
type fieldInputter struct {
	prompt prompter

	prefixFormat string

	encountered map[string]map[string]bool
	msg         *dynamic.Message
	fields      []*desc.FieldDescriptor
	dep         msgDep

	color prompt.Color

	isTopLevelMessage bool

	// enteredEmptyInput is used to terminate repeated field inputting
	// if input is empty and enteredEmptyInput is true, exit repeated input prompt
	enteredEmptyInput bool
}

type msgDep map[string]*desc.MessageDescriptor

// resolve dependencies of reqType
func resolveMessageDependency(msg *desc.MessageDescriptor, dep msgDep, encountered map[string]bool) {
	if encountered[msg.GetFullyQualifiedName()] {
		return
	}

	dep[msg.GetFullyQualifiedName()] = msg
	for _, f := range msg.GetFields() {
		if entity.IsMessageType(f.GetType()) {
			resolveMessageDependency(f.GetMessageType(), dep, encountered)
		}
	}
}

func newFieldInputter(prompter prompter, prefixFormat string, msg *dynamic.Message, msgType *desc.MessageDescriptor, isTopLevelMessage bool, color prompt.Color) *fieldInputter {
	dep := msgDep{}
	resolveMessageDependency(msgType, dep, map[string]bool{})
	return &fieldInputter{
		prompt:       prompter,
		prefixFormat: prefixFormat,
		encountered: map[string]map[string]bool{
			"oneof": map[string]bool{},
			"enum":  map[string]bool{},
		},
		msg:               msg,
		dep:               dep,
		isTopLevelMessage: isTopLevelMessage,
		color:             color,
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
func (i *fieldInputter) Input(fields []*desc.FieldDescriptor) (proto.Message, error) {
	if err := i.prompt.SetPrefixColor(i.color); err != nil {
		return nil, err
	}

	for _, field := range fields {
		if err := i.inputField(field); err != nil {
			return nil, err
		}

		if field.IsRepeated() {
			for {
				if err := i.inputField(field); err == EORF {
					break
				} else if err != nil {
					return nil, err
				}
			}
			i.enteredEmptyInput = false
		}
	}

	return i.msg, nil
}

func (i *fieldInputter) encounteredOneof(oneof *desc.OneOfDescriptor) bool {
	encountered := i.encountered["oneof"][oneof.GetFullyQualifiedName()]
	i.encountered["oneof"][oneof.GetFullyQualifiedName()] = true
	return encountered
}

func (i *fieldInputter) chooseOneof(oneof *desc.OneOfDescriptor) (*desc.FieldDescriptor, error) {
	options := make([]string, len(oneof.GetChoices()))
	descOf := map[string]*desc.FieldDescriptor{}
	for i, choice := range oneof.GetChoices() {
		options[i] = choice.GetName()
		descOf[choice.GetName()] = choice
	}

	choice, err := i.prompt.Select(oneof.GetName(), options)
	if err != nil {
		return nil, err
	}

	d, ok := descOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownOneofFieldName, choice)
	}

	return d, nil
}

func (i *fieldInputter) encounteredEnum(enum *desc.EnumDescriptor) bool {
	encountered := i.encountered["enum"][enum.GetFullyQualifiedName()]
	i.encountered["enum"][enum.GetFullyQualifiedName()] = true
	return encountered
}

func (i *fieldInputter) chooseEnum(enum *desc.EnumDescriptor) (*desc.EnumValueDescriptor, error) {
	options := make([]string, len(enum.GetValues()))
	descOf := map[string]*desc.EnumValueDescriptor{}
	for i, v := range enum.GetValues() {
		options[i] = v.GetName()
		descOf[v.GetName()] = v
	}

	choice, err := i.prompt.Select(enum.GetName(), options)
	if err != nil {
		return nil, err
	}

	d, ok := descOf[choice]
	if !ok {
		return nil, errors.Wrap(ErrUnknownEnumName, choice)
	}

	return d, nil
}

func (i *fieldInputter) inputField(field *desc.FieldDescriptor) error {
	// if oneof, choose one from selection
	if entity.IsOneOf(field) {
		oneof := field.GetOneOf()
		if i.encounteredOneof(oneof) {
			return nil
		}
		var err error
		field, err = i.chooseOneof(oneof)
		if err != nil {
			return err
		}
	}

	switch {
	case entity.IsEnumType(field):
		enum := field.GetEnumType()
		if i.encounteredEnum(enum) {
			return nil
		}
		v, err := i.chooseEnum(enum)
		if err != nil {
			return err
		}
		if err := i.setField(i.msg, field, v.GetNumber()); err != nil {
			return err
		}
	case entity.IsMessageType(field.GetType()):
		nestedFields := i.dep[field.GetMessageType().GetFullyQualifiedName()].GetFields()
		msgType := field.GetMessageType()
		msg, err := newFieldInputter(i.prompt, i.prefixFormat, dynamic.NewMessage(msgType), msgType, false, i.color).Input(nestedFields)
		if err != nil {
			return err
		}
		if err := i.setField(i.msg, field, msg); err != nil {
			return err
		}

		// increment prompt color to next one
		i.color = (i.color + 1) % 16
	default: // primitive type
		if err := i.prompt.SetPrefix(makePrefix(i.prefixFormat, field, i.isTopLevelMessage)); err != nil {
			return err
		}
		if err := i.inputPrimitiveField(i.msg, field); err != nil {
			return err
		}
	}
	return nil
}

func (i *fieldInputter) inputPrimitiveField(req *dynamic.Message, field *desc.FieldDescriptor) error {
	in := i.prompt.Input()

	if in == "" && field.IsRepeated() {
		if i.enteredEmptyInput {
			return EORF
		}
		i.enteredEmptyInput = true
		// ignore the input
		return i.inputPrimitiveField(req, field)
	}

	v, err := i.convertValue(in, field)
	if err != nil {
		return err
	}
	return i.setField(req, field, v)
}

func (i *fieldInputter) setField(req *dynamic.Message, field *desc.FieldDescriptor, v interface{}) error {
	if field.IsRepeated() {
		return req.TryAddRepeatedField(field, v)
	}

	return req.TrySetField(field, v)
}

// convertValue holds value and error of conversion
// each cast (Parse*) returns falsy value when failed to parse argument
func (i *fieldInputter) convertValue(pv string, f *desc.FieldDescriptor) (interface{}, error) {
	var v interface{}
	var err error

	switch f.GetType() {
	case descriptor.FieldDescriptorProto_TYPE_DOUBLE:
		v, err = strconv.ParseFloat(pv, 64)

	case descriptor.FieldDescriptorProto_TYPE_FLOAT:
		v, err = strconv.ParseFloat(pv, 32)
		v = float32(v.(float64))

	case descriptor.FieldDescriptorProto_TYPE_INT64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_UINT64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_INT32:
		v, err = strconv.ParseInt(pv, 10, 32)
		v = int32(v.(int64))

	case descriptor.FieldDescriptorProto_TYPE_UINT32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = uint32(v.(uint64))

	case descriptor.FieldDescriptorProto_TYPE_FIXED64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_FIXED32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = uint32(v.(uint64))

	case descriptor.FieldDescriptorProto_TYPE_BOOL:
		v, err = strconv.ParseBool(pv)

	case descriptor.FieldDescriptorProto_TYPE_STRING:
		// already string
		v = pv

	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		v = []byte(pv)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		v, err = strconv.ParseUint(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		v, err = strconv.ParseUint(pv, 10, 32)
		v = int32(v.(int64))

	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		v, err = strconv.ParseInt(pv, 10, 32)
		v = int32(v.(int64))

	default:
		return nil, fmt.Errorf("invalid type: %#v", f.GetType())
	}
	return v, err
}

// makePrefix makes prefix for field f.
// isTopLevelMessage is used to show passed f is a message and it is top-level message.
// for example, person field is a message, Person. and a part of BorrowBookRequest.
// also BorrowBookRequest is a top-level message.
//
// message BorrowBookRequest {
//  Person person = 1;
//   Book book = 2;
// }
//
func makePrefix(s string, f *desc.FieldDescriptor, isTopLevelMessage bool) string {
	ancestor := []string{}
	var d desc.Descriptor = f.GetParent()

	// if f is a top-level, message, exclude name of top-level message.
	if isTopLevelMessage {
		d = d.GetParent()
	}

	for d != nil {
		ancestor = append([]string{d.GetName()}, ancestor...)
		d = d.GetParent()
	}
	// remove file name
	ancestor = ancestor[1:]

	joinedAncestor := strings.Join(ancestor, "::")
	if joinedAncestor != "" {
		joinedAncestor += "::"
	}
	s = strings.Replace(s, "{ancestor}", joinedAncestor, -1)
	s = strings.Replace(s, "{name}", f.GetName(), -1)
	s = strings.Replace(s, "{type}", f.GetType().String(), -1)
	return s
}
