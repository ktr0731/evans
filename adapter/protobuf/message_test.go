package protobuf

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/ktr0731/evans/adapter/internal/protoparser"
	"github.com/ktr0731/evans/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		d := parseFile(t, []string{"message.proto"}, nil)
		assert.Len(t, d, 1)

		msgs := d[0].GetMessageTypes()
		assert.Len(t, msgs, 2)

		stringType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_STRING)]

		t.Run("Person message", func(t *testing.T) {
			personMsg := newMessage(msgs[0])
			assert.Equal(t, "Person", personMsg.Name())

			assert.Len(t, personMsg.Fields(), 1)
			personField := personMsg.Fields()[0]
			assert.Equal(t, personField.FieldName(), "name")

			assert.Equal(t, personField.PBType(), stringType)
		})

		t.Run("Nested message", func(t *testing.T) {
			nestedMsg := newMessage(msgs[1])
			assert.Equal(t, "Nested", nestedMsg.Name())

			assert.Len(t, nestedMsg.Fields(), 1)
			nestedMsgField := nestedMsg.Fields()[0]
			assert.Equal(t, nestedMsgField.FieldName(), "person")

			msgType := descriptor.FieldDescriptorProto_Type_name[int32(descriptor.FieldDescriptorProto_TYPE_MESSAGE)]
			assert.Equal(t, nestedMsgField.PBType(), msgType)

			assert.Equal(t, false, nestedMsgField.(entity.MessageField).IsCycled())
		})
	})

	t.Run("importing", func(t *testing.T) {
		libraryProto := testdata("importing", "library.proto")
		d, err := protoparser.ParseFile([]string{libraryProto}, nil)
		require.NoError(t, err)

		d = append(d, d[0].GetDependencies()...)
		assert.Len(t, d, 2)

		libMsgs := d[0].GetMessageTypes()
		bookMsgs := d[1].GetMessageTypes()

		assert.Equal(t, len(libMsgs)+len(bookMsgs), 4)
	})

	t.Run("self", func(t *testing.T) {
		d := parseFile(t, []string{"self.proto"}, nil)
		assert.Len(t, d, 1)

		msgs := d[0].GetMessageTypes()
		assert.Len(t, msgs, 3)

		msg := newMessage(msgs[0])
		assert.Equal(t, "Foo", msg.Name())
		assert.Len(t, msg.Fields(), 1)
		assert.True(t, msg.Fields()[0].(entity.MessageField).IsCycled())

		nextMsg := msg.Fields()[0].(entity.MessageField)
		assert.Len(t, nextMsg.Fields(), 1)
		assert.True(t, nextMsg.Fields()[0].(entity.MessageField).IsCycled())
	})

	t.Run("circular", func(t *testing.T) {
		d := parseFile(t, []string{"circular.proto"}, nil)
		assert.Len(t, d, 1)

		msgs := d[0].GetMessageTypes()
		assert.Len(t, msgs, 7)

		msgA, msgB := newMessage(msgs[0]), newMessage(msgs[1])
		assert.Equal(t, "A", msgA.Name())
		assert.Equal(t, "B", msgB.Name())
		assert.Len(t, msgA.Fields(), 1)
		assert.Len(t, msgB.Fields(), 1)
		assert.True(t, msgA.Fields()[0].(entity.MessageField).IsCycled())
		assert.True(t, msgB.Fields()[0].(entity.MessageField).IsCycled())

		t.Run("self-referenced", func(t *testing.T) {
			msg := newMessage(msgs[2])
			assert.Equal(t, "Foo", msg.Name())

			selfField := msg.Fields()[0].(entity.MessageField)
			assert.True(t, selfField.IsCycled())
			assert.Len(t, selfField.Fields(), 2)

			selfSelfField := selfField.Fields()[1].(entity.MessageField)
			assert.NotNil(t, selfSelfField.Fields())
			assert.True(t, selfSelfField.IsCycled())
		})

		t.Run("three messages", func(t *testing.T) {
			msgHoge, msgFuga, msgPiyo := newMessage(msgs[4]), newMessage(msgs[5]), newMessage(msgs[6])
			assert.Equal(t, "Hoge", msgHoge.Name())
			assert.Equal(t, "Fuga", msgFuga.Name())
			assert.Equal(t, "Piyo", msgPiyo.Name())
		})
	})
}
