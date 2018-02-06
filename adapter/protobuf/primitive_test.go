package protobuf

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/require"
)

func TestPrimitiveField(t *testing.T) {
	d := parseFile(t, "message.proto")
	msgs := d.GetMessageTypes()
	require.Len(t, msgs, 2)

	m := newMessage(msgs[0])
	require.Len(t, m.Fields(), 1)
	require.Equal(t, m.Fields()[0].FieldName(), "name")

	stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]
	require.Equal(t, m.Fields()[0].Type(), stringType)
}
