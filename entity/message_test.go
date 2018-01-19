package entity

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	d := parseFile(t, "message.proto")
	msgs := d.GetMessageTypes()
	require.Len(t, msgs, 2)

	personMsg := newMessage(msgs[0])
	assert.Equal(t, "Person", personMsg.Name())
	assert.Equal(t, NON_FIELD, personMsg.Number())

	require.Len(t, personMsg.Fields, 1)
	personField := personMsg.Fields[0]
	assert.Equal(t, personField.Name(), "name")
	assert.Equal(t, personField.Number(), int32(1))

	stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]
	assert.Equal(t, personField.Type(), stringType)

	nestedMsg := newMessage(msgs[1])
	assert.Equal(t, "Nested", nestedMsg.Name())
	assert.Equal(t, NON_FIELD, nestedMsg.Number())

	require.Len(t, nestedMsg.Fields, 1)
	nestedMsgField := nestedMsg.Fields[0]
	assert.Equal(t, nestedMsgField.Name(), "person")
	assert.Equal(t, nestedMsgField.Number(), int32(1))

	msgType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_MESSAGE)]
	assert.Equal(t, nestedMsgField.Type(), msgType)

	// person
	m, err := nestedMsgField.(*Message)
	require.True(t, err)

	require.Len(t, m.Fields, 1)
	mField := m.Fields[0]
	assert.Equal(t, mField.Name(), "name")
	assert.Equal(t, mField.Number(), int32(1))
	assert.Equal(t, mField.Type(), stringType)
}
