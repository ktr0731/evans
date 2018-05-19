package protobuf

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		d := parseFile(t, []string{"message.proto"}, nil)
		require.Len(t, d, 1)

		msgs := d[0].GetMessageTypes()
		require.Len(t, msgs, 2)

		stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]

		t.Run("Person message", func(t *testing.T) {
			personMsg := newMessage(msgs[0])
			require.Equal(t, "Person", personMsg.Name())

			require.Len(t, personMsg.Fields(), 1)
			personField := personMsg.Fields()[0]
			require.Equal(t, personField.FieldName(), "name")

			require.Equal(t, personField.PBType(), stringType)
		})

		t.Run("Nested message", func(t *testing.T) {
			nestedMsg := newMessage(msgs[1])
			require.Equal(t, "Nested", nestedMsg.Name())

			require.Len(t, nestedMsg.Fields(), 1)
			nestedMsgField := nestedMsg.Fields()[0]
			require.Equal(t, nestedMsgField.FieldName(), "person")

			msgType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_MESSAGE)]
			require.Equal(t, nestedMsgField.PBType(), msgType)
		})
	})

	t.Run("importing", func(t *testing.T) {
		libraryProto := testdata("importing", "library.proto")
		d, err := protoparser.ParseFile([]string{libraryProto}, nil)
		require.NoError(t, err)

		d = append(d, d[0].GetDependencies()...)
		require.Len(t, d, 2)

		libMsgs := d[0].GetMessageTypes()
		bookMsgs := d[1].GetMessageTypes()

		assert.Equal(t, len(libMsgs)+len(bookMsgs), 4)
	})

	t.Run("self", func(t *testing.T) {
		d := parseFile(t, []string{"self.proto"}, nil)
		require.Len(t, d, 1)

		msgs := d[0].GetMessageTypes()
		require.Len(t, msgs, 1)

		msg := newMessage(msgs[0])
		require.Equal(t, "Foo", msg.Name())
	})
}
