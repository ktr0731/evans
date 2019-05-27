package proto

import (
	"fmt"
	"strconv"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/pkg/errors"
)

var protoDefaults = map[descriptor.FieldDescriptorProto_Type]interface{}{
	descriptor.FieldDescriptorProto_TYPE_DOUBLE:   float64(0),
	descriptor.FieldDescriptorProto_TYPE_FLOAT:    float32(0),
	descriptor.FieldDescriptorProto_TYPE_INT64:    int64(0),
	descriptor.FieldDescriptorProto_TYPE_UINT64:   uint64(0),
	descriptor.FieldDescriptorProto_TYPE_INT32:    int32(0),
	descriptor.FieldDescriptorProto_TYPE_UINT32:   uint32(0),
	descriptor.FieldDescriptorProto_TYPE_FIXED64:  uint64(0),
	descriptor.FieldDescriptorProto_TYPE_FIXED32:  uint32(0),
	descriptor.FieldDescriptorProto_TYPE_BOOL:     false,
	descriptor.FieldDescriptorProto_TYPE_STRING:   "",
	descriptor.FieldDescriptorProto_TYPE_BYTES:    []byte{},
	descriptor.FieldDescriptorProto_TYPE_SFIXED64: int64(0),
	descriptor.FieldDescriptorProto_TYPE_SFIXED32: int32(0),
	descriptor.FieldDescriptorProto_TYPE_SINT64:   int64(0),
	descriptor.FieldDescriptorProto_TYPE_SINT32:   int32(0),
}

// convertValue converts a string input pv to fieldType.
func convertValue(pv string, fieldType descriptor.FieldDescriptorProto_Type) (interface{}, error) {
	if pv == "" {
		d, found := protoDefaults[fieldType]
		if found {
			return d, nil
		}
		// if not found, we'll let the normal code execute
	}

	var v interface{}
	var err error

	switch fieldType {
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
	// His expects "ogiso" in string, but backslashes in the input are not interpreted as an escape sequence.
	// So, we need to call strconv.Unquote to interpret backslashes as an escape sequence.
	case descriptor.FieldDescriptorProto_TYPE_BYTES:
		pv, err = strconv.Unquote(`"` + pv + `"`)
		v = []byte(pv)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SFIXED32:
		v, err = strconv.ParseInt(pv, 10, 32)
		v = int32(v.(int64))

	case descriptor.FieldDescriptorProto_TYPE_SINT64:
		v, err = strconv.ParseInt(pv, 10, 64)

	case descriptor.FieldDescriptorProto_TYPE_SINT32:
		v, err = strconv.ParseInt(pv, 10, 32)
		v = int32(v.(int64))

	default:
		return nil, fmt.Errorf("invalid type: %s", fieldType)
	}
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert an inputted value '%s' to type %s", pv, fieldType)
	}
	return v, nil
}
