package protobuf

import (
	"errors"
	"fmt"
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
		return nil, fmt.Errorf("invalid type: %s", t)
	}
	return v, err
}

// ToEntitiesFrom normalizes descriptors to entities
//
// package
// ├ messages
// ├ enums
// └ services
//   └ rpcs
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
