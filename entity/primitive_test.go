package entity

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
)

func TestPrimitiveField(t *testing.T) {
	d := parseFile(t, "message.proto")
	msgs := d.GetMessageTypes()
	assert.Len(t, msgs, 2)

	m := newMessage(msgs[0])
	assert.Equal(t, m.Name(), "Person")
	assert.Len(t, m.Fields(), 1)
	assert.Equal(t, m.Fields()[0].Name, "name")

	stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]
	assert.Equal(t, m.Fields()[0].Type, stringType)
}
