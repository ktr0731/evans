package env

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	prompt "github.com/c-bata/go-prompt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

// field is used to read and store input for each field
// if field type is message, this struct is recursive
type field struct {
	isPrimitive bool
	name        string
	pVal        *string
	mVal        []*field
	descType    descriptor.FieldDescriptorProto_Type
	desc        *desc.FieldDescriptor
}

// Call calls a RPC which is selected
// RPC is called after inputting field values interactively
func (e *Env) Call(name string) (string, error) {
	rpc, err := e.GetRPC(name)
	if err != nil {
		return "", err
	}

	input, err := e.inputFields([]string{}, rpc.RequestType.GetFields(), prompt.DarkGreen)
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
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", e.config.Server.Host, e.config.Server.Port), grpc.WithInsecure())
	if err != nil {
		return "", err
	}
	defer conn.Close()

	ep := e.genEndpoint(name)
	if err := grpc.Invoke(context.Background(), ep, req, res, conn); err != nil {
		return "", err
	}

	m := jsonpb.Marshaler{Indent: "  "}
	json, err := m.MarshalToString(res)
	if err != nil {
		return "", err
	}

	return json + "\n", nil
}

func (e *Env) genEndpoint(rpcName string) string {
	ep := fmt.Sprintf("/%s.%s/%s", e.state.currentPackage, e.state.currentService, rpcName)
	return ep
}

func (e *Env) setInput(req *dynamic.Message, fields []*field) error {
	for _, f := range fields {
		if !f.isPrimitive {
			// TODO
			msg := dynamic.NewMessage(f.desc.GetMessageType())
			if err := e.setInput(msg, f.mVal); err != nil {
				return err
			}
			req.SetField(f.desc, msg)
		} else {
			pv := *f.pVal

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
				v, err = strconv.ParseInt(*f.pVal, 10, 32)
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
				return fmt.Errorf("invalid type: %#v", f.descType)
			}

			if err != nil {
				return err
			}
			if err := req.TrySetField(f.desc, v); err != nil {
				return err
			}

		}
	}
	return nil
}

func (e *Env) inputFields(ancestor []string, fields []*desc.FieldDescriptor, color prompt.Color) ([]*field, error) {
	input := make([]*field, len(fields))
	max := maxLen(fields, e.config.InputPromptFormat)
	for i, f := range fields {
		input[i] = &field{
			name:     f.GetName(),
			desc:     f,
			descType: f.GetType(),
		}

		// message field or primitive field
		if isMessageType(f.GetType()) {
			fields, err := e.inputFields(append(ancestor, f.GetName()), f.GetMessageType().GetFields(), color)
			if err != nil {
				return nil, errors.Wrap(err, "failed to read inputs")
			}
			input[i].mVal = fields
			color++
		} else {
			promptFormat := e.config.InputPromptFormat
			ancestor := strings.Join(ancestor, e.config.AncestorDelimiter)
			if ancestor != "" {
				ancestor = "@" + ancestor
			}
			promptFormat = strings.Replace(promptFormat, "{ancestor}", ancestor, -1)
			promptFormat = strings.Replace(promptFormat, "{name}", f.GetName(), -1)
			promptFormat = strings.Replace(promptFormat, "{type}", f.GetType().String(), -1)

			l := prompt.Input(
				fmt.Sprintf("%"+strconv.Itoa(max)+"s", promptFormat),
				inputCompleter,
				prompt.OptionPrefixTextColor(color),
			)
			input[i].isPrimitive = true
			input[i].pVal = &l
		}
	}
	return input, nil
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

func inputCompleter(d prompt.Document) []prompt.Suggest {
	return nil
}
