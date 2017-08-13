package model

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/stretchr/testify/assert"
)

func TestNewFields(t *testing.T) {
	const pkgName = "steinsgate"
	desc := fileDesc(t, []string{"testdata/test.proto"}, []string{})
	tests := map[string]struct {
		msgName string
		expect  []*Field
		err     error
	}{
		"normal": {
			msgName: "Person", err: nil,
			expect: []*Field{
				{Name: "name", Type: descriptor.FieldDescriptorProto_TYPE_STRING},
			}},
		"nested": {
			msgName: "TimeleapReq", err: nil,
			expect: []*Field{
				{Name: "when", Type: descriptor.FieldDescriptorProto_TYPE_STRING},
				{Name: "person", Type: descriptor.FieldDescriptorProto_TYPE_MESSAGE, IsMessage: true},
			}},
	}

	for title, test := range tests {
		t.Run(title, func(t *testing.T) {
			msg := getMessage(t, desc, pkgName, test.msgName)

			// Message struct is used only extract field from descriptor
			actual, err := NewFields(func(msgName string) (*Message, error) {
				return getMessage(t, desc, pkgName, msgName), nil
			}, &Message{Desc: msg.Desc})

			// Erase descriptor because it not need
			for _, a := range actual {
				a.Desc = nil
				a.Fields = nil
			}

			assert.Equal(t, test.err, err)
			assert.Exactly(t, test.expect, actual)
		})
	}
}
