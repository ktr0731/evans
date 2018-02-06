package protobuf

import (
	"path/filepath"
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		d := parseFile(t, "message.proto")
		msgs := d.GetMessageTypes()
		require.Len(t, msgs, 2)

		stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]

		t.Run("Person message", func(t *testing.T) {
			personMsg := newMessage(msgs[0])
			assert.Equal(t, "Person", personMsg.Name())
			assert.Equal(t, NON_FIELD, personMsg.Number())

			require.Len(t, personMsg.Fields(), 1)
			personField := personMsg.Fields()[0]
			assert.Equal(t, personField.Name, "name")

			assert.Equal(t, personField.Type, stringType)
		})

		t.Run("Nested message", func(t *testing.T) {
			nestedMsg := newMessage(msgs[1])
			assert.Equal(t, "Nested", nestedMsg.Name())
			assert.Equal(t, NON_FIELD, nestedMsg.Number())

			require.Len(t, nestedMsg.Fields(), 1)
			nestedMsgField := nestedMsg.Fields()[0]
			assert.Equal(t, nestedMsgField.Name, "person")

			msgType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_MESSAGE)]
			assert.Equal(t, nestedMsgField.Type, msgType)
		})
	})

	t.Run("importing", func(t *testing.T) {
		libraryProto := filepath.Join("importing", "library.proto")
		d := parseDependFiles(t, libraryProto, filepath.Join("testdata", "importing"))

		require.Len(t, d, 2)

		bookMsgs := d[0].GetMessageTypes()
		libraryMsgs := d[1].GetMessageTypes()

		require.Equal(t, len(bookMsgs)+len(libraryMsgs), 4)
	})

	t.Run("self", func(t *testing.T) {
		d := parseFile(t, "self.proto")
		msgs := d.GetMessageTypes()
		require.Len(t, msgs, 1)

		msg := newMessage(msgs[0])
		assert.Equal(t, "Foo", msg.Name())
	})
}
