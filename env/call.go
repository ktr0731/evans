package env

import (
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"github.com/AlecAivazis/survey"
	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/config"
	"github.com/pkg/errors"
)

// fieldable is only used to set primitive, enum, oneof fields.
type fieldable interface {
	fieldable()
	isNil() bool
}

type baseField struct {
	name     string
	descType descriptor.FieldDescriptorProto_Type
	desc     *desc.FieldDescriptor
}

func newBaseField(f *desc.FieldDescriptor) *baseField {
	return &baseField{
		name:     f.GetName(),
		desc:     f,
		descType: f.GetType(),
	}
}

func (f *baseField) fieldable() {}

// primitiveField is used to read and store input for each primitiveField
type primitiveField struct {
	*baseField
	val string
}

func (f *primitiveField) isNil() bool {
	return f.val == ""
}

type messageField struct {
	*baseField
	val []fieldable
}

func (f *messageField) isNil() bool {
	return f.val == nil
}

type repeatedField struct {
	*baseField
	val []fieldable
}

func (f *repeatedField) isNil() bool {
	return len(f.val) == 0
}

type enumField struct {
	*baseField
	val *desc.EnumValueDescriptor
}

func (f *enumField) isNil() bool {
	return f.val == nil
}

// Call calls a RPC which is selected
// RPC is called after inputting field values interactively
func (e *Env) Call(name string) (string, error) {
	rpc, err := e.GetRPC(name)
	if err != nil {
		return "", err
	}

	// TODO: GetFields は OneOf の要素まで取得してしまう
	input, err := e.inputFields([]string{}, rpc.RequestType, prompt.DarkGreen)
	if errors.Cause(err) == io.EOF {
		return "", nil
	} else if err != nil {
		return "", err
	}

	req := dynamic.NewMessage(rpc.RequestType)
	if err = e.setInput(req, input); err != nil {
		return "", err
	}

	res := dynamic.NewMessage(rpc.ResponseType)
	conn, err := connect(e.config.Server)
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ep := e.genEndpoint(name)
	if err := grpc.Invoke(context.Background(), ep, req, res, conn); err != nil {
		return "", err
	}

	out, err := formatOutput(res)
	if err != nil {
		return "", err
	}

	return out, nil
}

func (e *Env) CallWithScript(input io.Reader, rpcName string) error {
	rpc, err := e.GetRPC(rpcName)
	if err != nil {
		return err
	}
	req := dynamic.NewMessage(rpc.RequestType)
	if err := jsonpb.Unmarshal(input, req); err != nil {
		return err
	}
	res := dynamic.NewMessage(rpc.ResponseType)
	conn, err := connect(e.config.Server)
	if err != nil {
		return err
	}
	defer conn.Close()

	if err := grpc.Invoke(context.Background(), e.genEndpoint(rpcName), req, res, conn); err != nil {
		return err
	}

	out, err := formatOutput(res)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, out)

	return nil
}

func (e *Env) genEndpoint(rpcName string) string {
	ep := fmt.Sprintf("/%s.%s/%s", e.state.currentPackage, e.state.currentService, rpcName)
	return ep
}

func (e *Env) setInput(req *dynamic.Message, fields []fieldable) error {
	for _, field := range fields {
		switch f := field.(type) {
		case *primitiveField:
			pv := f.val

			v, err := castPrimitiveType(f, pv)
			if err != nil {
				return err
			}
			if err := req.TrySetField(f.desc, v); err != nil {
				return err
			}

		case *messageField:
			// TODO
			msg := dynamic.NewMessage(f.desc.GetMessageType())
			if err := e.setInput(msg, f.val); err != nil {
				return err
			}
			req.SetField(f.desc, msg)
		case *repeatedField:
			// ここの f.desc に Add する

			if f.desc.GetMessageType() != nil {
				msg := dynamic.NewMessage(f.desc.GetMessageType())
				if err := e.setInput(msg, f.val); err != nil {
					return err
				}
				req.TryAddRepeatedField(f.desc, msg)
			} else { // primitive type
				for _, field := range f.val {
					f2 := field.(*primitiveField)
					v, err := castPrimitiveType(f2, f2.val)
					if err != nil {
						return err
					}
					if err := req.TryAddRepeatedField(f.desc, v); err != nil {
						return err
					}
				}
			}
		case *enumField:
			// TODO
			req.SetField(f.desc, f.val.GetNumber())
		}
	}
	return nil
}

func (e *Env) inputFields(ancestor []string, msg *desc.MessageDescriptor, color prompt.Color) ([]fieldable, error) {
	fields := msg.GetFields()

	input := make([]fieldable, 0, len(fields))
	max := maxLen(fields, e.config.InputPromptFormat)
	// TODO: ずれてる
	promptFormat := fmt.Sprintf("%"+strconv.Itoa(max)+"s", e.config.InputPromptFormat)

	inputField := e.fieldInputer(ancestor, promptFormat, color)

	encountered := map[string]map[string]bool{
		"oneof": map[string]bool{},
		"enum":  map[string]bool{},
	}
	for _, f := range fields {
		var in fieldable
		// message field, enum field or primitive field
		switch {
		case isOneOf(f):
			oneOf := f.GetOneOf()

			if encountered["oneof"][oneOf.GetFullyQualifiedName()] {
				continue
			}

			encountered["oneof"][oneOf.GetFullyQualifiedName()] = true

			opts := make([]string, len(oneOf.GetChoices()))
			optMap := map[string]*desc.FieldDescriptor{}
			for i, c := range oneOf.GetChoices() {
				opts[i] = c.GetName()
				optMap[c.GetName()] = c
			}

			var choice string
			err := survey.AskOne(&survey.Select{
				Message: oneOf.GetName(),
				Options: opts,
			}, &choice, nil)
			if err != nil {
				return nil, err
			}

			f = optMap[choice]
		case isEnumType(f):
			enum := f.GetEnumType()
			if encountered["enum"][enum.GetFullyQualifiedName()] {
				continue
			}

			encountered["enum"][enum.GetFullyQualifiedName()] = true

			opts := make([]string, len(enum.GetValues()))
			optMap := map[string]*desc.EnumValueDescriptor{}
			for i, o := range enum.GetValues() {
				opts[i] = o.GetName()
				optMap[o.GetName()] = o
			}

			var choice string
			err := survey.AskOne(&survey.Select{
				Message: enum.GetName(),
				Options: opts,
			}, &choice, nil)
			if err != nil {
				return nil, err
			}
			in = &enumField{
				baseField: newBaseField(f),
				val:       optMap[choice],
			}
		}

		if f.IsRepeated() {
			var repeated []fieldable
			// TODO: repeated であることを prompt に出したい
			for {
				s, err := inputField(f)
				if err != nil {
					return nil, err
				}
				if s.isNil() {
					break
				}
				repeated = append(repeated, s)
			}
			in = &repeatedField{
				baseField: newBaseField(f),
				val:       repeated,
			}
		} else if !isEnumType(f) {
			var err error
			in, err = inputField(f)
			if err != nil {
				return nil, err
			}
		}

		input = append(input, in)
	}
	return input, nil
}

func connect(config *config.Server) (*grpc.ClientConn, error) {
	// TODO: connection を使いまわしたい
	return grpc.Dial(fmt.Sprintf("%s:%s", config.Host, config.Port), grpc.WithInsecure())
}

func formatOutput(input proto.Message) (string, error) {
	m := jsonpb.Marshaler{Indent: "  "}
	out, err := m.MarshalToString(input)
	if err != nil {
		return "", err
	}
	return out + "\n", nil
}

// fieldInputer let us enter primitive or message field.
func (e *Env) fieldInputer(ancestor []string, promptFormat string, color prompt.Color) func(*desc.FieldDescriptor) (fieldable, error) {
	return func(f *desc.FieldDescriptor) (fieldable, error) {
		if isMessageType(f.GetType()) {
			fields, err := e.inputFields(append(ancestor, f.GetName()), f.GetMessageType(), color)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read inputs")
			}
			color = prompt.DarkGreen + (color+1)%16
			return &messageField{
				baseField: newBaseField(f),
				val:       fields,
			}, nil
		} else { // primitive
			promptStr := promptFormat
			ancestor := strings.Join(ancestor, e.config.AncestorDelimiter)
			if ancestor != "" {
				ancestor = "@" + ancestor
			}
			// TODO: text template
			promptStr = strings.Replace(promptStr, "{ancestor}", ancestor, -1)
			promptStr = strings.Replace(promptStr, "{name}", f.GetName(), -1)
			promptStr = strings.Replace(promptStr, "{type}", f.GetType().String(), -1)

			in := prompt.Input(
				promptStr,
				inputCompleter,
				prompt.OptionPrefixTextColor(color),
			)

			return &primitiveField{
				baseField: newBaseField(f),
				val:       in,
			}, nil
		}
	}
}

func castPrimitiveType(f *primitiveField, pv string) (interface{}, error) {
	// it holds value and error of conversion
	// each cast (Parse*) returns falsy value when failed to parse argument
	var v interface{}
	var err error

	switch f.descType {
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
		v, err = strconv.ParseInt(f.val, 10, 32)
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
		return nil, fmt.Errorf("invalid type: %#v", f.descType)
	}
	return v, err
}

func maxLen(fields []*desc.FieldDescriptor, format string) int {
	var max int
	for _, f := range fields {
		if isMessageType(f.GetType()) {
			continue
		}
		prompt := format
		elems := map[string]string{
			"name": f.GetName(),
			"type": f.GetType().String(),
		}
		for k, v := range elems {
			prompt = strings.Replace(prompt, "{"+k+"}", v, -1)
		}
		l := len(format)
		if l > max {
			max = l
		}
	}
	return max
}

func isMessageType(typeName descriptor.FieldDescriptorProto_Type) bool {
	return typeName == descriptor.FieldDescriptorProto_TYPE_MESSAGE
}

func isOneOf(f *desc.FieldDescriptor) bool {
	return f.GetOneOf() != nil
}

func isEnumType(f *desc.FieldDescriptor) bool {
	return f.GetEnumType() != nil
}

func inputCompleter(d prompt.Document) []prompt.Suggest {
	return nil
}
