package gateway

import (
	"fmt"
	"strconv"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

type PromptInputter struct {
	env entity.Environment
}

func NewPromptInputter(env entity.Environment) *PromptInputter {
	return &PromptInputter{env}
}

// TODO: 対象メッセージが依存しているメッセージも含めて与えてやる
func (i *PromptInputter) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	req := dynamic.NewMessage(reqType)
	fields := reqType.GetFields()

	if err := newFieldInputter(req, fields).Input(); err != nil {
		return nil, err
	}
	return req, nil
}

// fieldInputter inputs each fields of req in interactively
// first fieldInputter is instantiated per one request
// fieldInputter will instantiate another fieldInputter instance for nested messages
//
// e.g.
// message Foo {
// 	Bar bar = 1;
//	string baz = 2;
// }
//
// it generate two fieldInputter
// one is used for Foo's primitive fields
// another one is used for bar's primitive fields
type fieldInputter struct {
	encountered map[string]map[string]bool
	req         *dynamic.Message
	fields      []*desc.FieldDescriptor
}

func newFieldInputter(req *dynamic.Message, fields []*desc.FieldDescriptor) *fieldInputter {
	return &fieldInputter{
		encountered: map[string]map[string]bool{
			"oneof": map[string]bool{},
			"enum":  map[string]bool{},
		},
		req:    req,
		fields: fields,
	}
}

func (i *fieldInputter) Input() error {
	for _, field := range i.fields {
		switch {
		case entity.IsOneOf(field):
			oneof := field.GetOneOf()
			if i.encounteredOneof(oneof) {
				continue
			}
			v, err := i.chooseOneof(oneof)
			if err != nil {
				return err
			}
			i.req.TrySetField(field, v)
		case entity.IsEnumType(field):
			enum := field.GetEnumType()
			if i.encounteredEnum(enum) {
				continue
			}
			v, err := i.chooseEnum(enum)
			if err != nil {
				return err
			}
			i.req.SetField(field, v.GetNumber())
		case entity.IsMessageType(field.GetType()):
			// message を解決する必要がある
			// newFieldInputter(i.req, field).Input()
		}
	}
	return nil
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
	// TODO:
	in := prompt.Input(
		"test>",
		nil,
		nil,
	)

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
