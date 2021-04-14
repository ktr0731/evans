package proto

import (
	"reflect"
	"testing"

	"google.golang.org/protobuf/types/descriptorpb"
)

func Test_convertValue(t *testing.T) {
	cases := map[string]struct {
		v         string
		fieldType descriptorpb.FieldDescriptorProto_Type

		expected interface{}
		hasErr   bool
	}{
		"default of string": {
			v:         "",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_STRING,
			expected:  "",
		},
		"double": {
			v:         "100.2",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_DOUBLE,
			expected:  float64(100.2),
		},
		"float": {
			v:         "100.2",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_FLOAT,
			expected:  float32(100.2),
		},
		"int64": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_INT64,
			expected:  int64(100),
		},
		"uint64": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_UINT64,
			expected:  uint64(100),
		},
		"int32": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_INT32,
			expected:  int32(100),
		},
		"uint32": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_UINT32,
			expected:  uint32(100),
		},
		"fixed64": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_FIXED64,
			expected:  uint64(100),
		},
		"fixed32": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_FIXED32,
			expected:  uint32(100),
		},
		"bool": {
			v:         "true",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_BOOL,
			expected:  true,
		},
		"string": {
			v:         "violet evergarden",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_STRING,
			expected:  "violet evergarden",
		},
		"bytes": {
			v:         "ogiso",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_BYTES,
			expected:  []byte("ogiso"),
		},
		"bytes (non-ascii string)": {
			v:         "小木曽",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_BYTES,
			expected:  []byte("小木曽"),
		},
		"bytes (Unicode literals)": {
			v:         "\u5c0f\u6728\u66fd",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_BYTES,
			expected:  []byte("小木曽"),
		},
		"sfixed64": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_SFIXED64,
			expected:  int64(100),
		},
		"sfixed32": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_SFIXED32,
			expected:  int32(100),
		},
		"sint64": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_SINT64,
			expected:  int64(100),
		},
		"sint32": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_SINT32,
			expected:  int32(100),
		},
		"invalid type": {
			v:         "100",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_SINT64 + 1, // Invalid type.
			hasErr:    true,
		},
		"invalid value": {
			v:         "100.10",
			fieldType: descriptorpb.FieldDescriptorProto_TYPE_INT32,
			hasErr:    true,
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			actual, err := convertValue(c.v, c.fieldType)
			if c.hasErr {
				if err == nil {
					t.Errorf("convertValue must return an error, but got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("convertValue must not return errors, but got an error: '%s'", err)
			}

			if !reflect.DeepEqual(c.expected, actual) {
				t.Errorf("expected '%v' (type = %T), but got '%v' (type = %T)",
					c.expected, c.expected, actual, actual)
			}
		})
	}
}
