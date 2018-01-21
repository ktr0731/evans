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
)

// for mocking
type prompter interface {
	Input() string
	SetPrefix(prefix string) error
}

type RealPrompter struct {
	*prompt.Prompt
}

func (p *RealPrompter) Input() string {
	return p.Prompt.Input()
}

func (p *RealPrompter) SetPrefix(prefix string) error {
	return prompt.OptionPrefix(prefix)(p.Prompt)
}

type PromptInputter struct {
	*promptInputter
}

// mixin go-prompt
func NewPromptInputter(config *config.Config, env entity.Environment) *PromptInputter {
	executor := func(in string) {
		return
	}
	completer := func(d prompt.Document) []prompt.Suggest {
		return nil
	}
	return &PromptInputter{newPromptInputter(&RealPrompter{prompt.New(executor, completer)}, config, env)}
}

type promptInputter struct {
	prompt prompter
	config *config.Config
	env    entity.Environment
}

func newPromptInputter(prompt prompter, config *config.Config, env entity.Environment) *promptInputter {
	return &promptInputter{
		prompt: prompt,
		config: config,
		env:    env,
	}
}

func (i *promptInputter) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	req := dynamic.NewMessage(reqType)
	fields := reqType.GetFields()

	return newFieldInputter(i.prompt, i.config.Env.InputPromptFormat, req, reqType, true).Input(fields)
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
type fieldInputter struct {
	prompt prompter

	prefixFormat string

	encountered map[string]map[string]bool
	msg         *dynamic.Message
	fields      []*desc.FieldDescriptor
	dep         messageDependency

	isTopLevelMessage bool
}

type messageDependency map[string]*desc.MessageDescriptor

// resolve dependencies of reqType
func resolveMessageDependency(msg *desc.MessageDescriptor, dep messageDependency, encountered map[string]bool) {
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

func newFieldInputter(prompt prompter, prefixFormat string, msg *dynamic.Message, msgType *desc.MessageDescriptor, isTopLevelMessage bool) *fieldInputter {
	dep := messageDependency{}
	resolveMessageDependency(msgType, dep, map[string]bool{})
	return &fieldInputter{
		prompt:       prompt,
		prefixFormat: prefixFormat,
		encountered: map[string]map[string]bool{
			"oneof": map[string]bool{},
			"enum":  map[string]bool{},
		},
		msg:               msg,
		dep:               dep,
		isTopLevelMessage: isTopLevelMessage,
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
	for _, field := range fields {
		switch {
		case entity.IsOneOf(field):
			oneof := field.GetOneOf()
			if i.encounteredOneof(oneof) {
				continue
			}
			v, err := i.chooseOneof(oneof)
			if err != nil {
				return nil, err
			}
			if err := i.msg.TrySetField(field, v); err != nil {
				return nil, err
			}
		case entity.IsEnumType(field):
			enum := field.GetEnumType()
			if i.encounteredEnum(enum) {
				continue
			}
			v, err := i.chooseEnum(enum)
			if err != nil {
				return nil, err
			}
			if err := i.msg.TrySetField(field, v.GetNumber()); err != nil {
				return nil, err
			}
		case entity.IsMessageType(field.GetType()):
			nestedFields := i.dep[field.GetMessageType().GetFullyQualifiedName()].GetFields()
			msgType := field.GetMessageType()
			msg, err := newFieldInputter(i.prompt, i.prefixFormat, dynamic.NewMessage(msgType), msgType, false).Input(nestedFields)
			if err != nil {
				return nil, err
			}
			if err := i.msg.TrySetField(field, msg); err != nil {
				return nil, err
			}
		default: // primitive type
			if err := i.prompt.SetPrefix(makePrefix(i.prefixFormat, field, i.isTopLevelMessage)); err != nil {
				return nil, err
			}
			if err := i.inputField(i.msg, field); err != nil {
				return nil, err
			}
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

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: oneof.GetName(),
		Options: options,
	}, &choice, nil)
	if err != nil {
		return nil, err
	}

	return descOf[choice], nil
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

	var choice string
	err := survey.AskOne(&survey.Select{
		Message: enum.GetName(),
		Options: options,
	}, &choice, nil)
	if err != nil {
		return nil, err
	}

	return descOf[choice], nil
}

func (i *fieldInputter) inputField(req *dynamic.Message, field *desc.FieldDescriptor) error {
	in := i.prompt.Input()

	v, err := i.convertValue(in, field)
	if err != nil {
		return err
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
