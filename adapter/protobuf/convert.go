package protobuf

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/ktr0731/evans/entity"
)

// ConvertValue holds value and error of conversion
// each cast (Parse*) returns falsy value when failed to parse argument
func ConvertValue(pv string, field entity.PrimitiveField) (interface{}, error) {
	f, ok := field.(*primitiveField)
	if !ok {
		return nil, errors.New("type assertion failed")
	}

	var v interface{}
	var err error

	t := descriptor.FieldDescriptorProto_Type(descriptor.FieldDescriptorProto_Type_value[f.PBType()])
	switch t {
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

	// Use strconv.Unquote to interpret byte literals and Unicode literals.
	// For example, a user inputs `\x6f\x67\x69\x73\x6f`,
	// His expects "ogiso" in string, but backslashes in the input are
	// not interpreted as an escape sequence.
	// So, we need to call strconv.Unquote to interpret backslashes as an escape sequence.
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		pv, err = strconv.Unquote(`"` + pv + `"`)
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
		return nil, fmt.Errorf("invalid type: %s", t)
	}
	return v, err
}

// ToEntitiesFrom normalizes descriptors to entities
//
// package
// - messages
// - enums
// - services
//   - rpcs
//
func ToEntitiesFrom(files []*desc.FileDescriptor) ([]*entity.Package, error) {
	pkgNames := map[string]bool{}
	msgMap := map[string][]entity.Message{}
	svcMap := map[string][]entity.Service{}
	for _, f := range files {
		pkgName := f.GetPackage()

		pkgNames[pkgName] = true

		for _, msg := range f.GetMessageTypes() {
			msgMap[pkgName] = append(msgMap[pkgName], newMessage(msg))
		}
		for _, svc := range f.GetServices() {
			svcMap[pkgName] = append(svcMap[pkgName], newService(svc))
		}
	}

	var pkgs []*entity.Package
	for pkgName := range pkgNames {
		pkgs = append(pkgs, entity.NewPackage(pkgName, msgMap[pkgName], svcMap[pkgName]))
	}

	return pkgs, nil
}

// ToEntitiesFromServiceDescriptors normalizes service descriptors to entities.
// this API is called if the target server has enabled gRPC reflection.
// reflection service returns available services, so Evans needs to convert to entities
// from service descriptors, not file descriptors.
// Also ToEntitiesFromServiceDescriptors returns messages which is containing to request/response message fields or itself.
func ToEntitiesFromServiceDescriptors(services []*desc.ServiceDescriptor) ([]entity.Service, []entity.Message) {
	msgs := make([]entity.Message, 0, 2) // request, response messages
	svcs := make([]entity.Service, 0, len(services))

	encounteredMessage := map[string]bool{}
	for _, s := range services {
		svc := newService(s)
		svcs = append(svcs, svc)

		for _, rpc := range svc.RPCs() {
			for _, msg := range []entity.Message{rpc.RequestMessage(), rpc.ResponseMessage()} {
				for _, f := range msg.Fields() {
					if f.Type() != entity.FieldTypeMessage {
						continue
					}
					mf := f.(*messageField)
					if !encounteredMessage[mf.Message.Name()] {
						msgs = append(msgs, mf.Message)
						encounteredMessage[mf.Message.Name()] = true
					}
				}
				if !encounteredMessage[msg.Name()] {
					msgs = append(msgs, msg)
					encounteredMessage[msg.Name()] = true
				}
			}
		}
	}
	sort.Slice(svcs, func(i, j int) bool {
		return svcs[i].Name() < svcs[j].Name()
	})
	sort.Slice(msgs, func(i, j int) bool {
		return msgs[i].Name() < msgs[j].Name()
	})
	return svcs, msgs
}
