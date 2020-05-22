package proto

import (
	"testing"

	"github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/ktr0731/evans/fill"
	"github.com/ktr0731/evans/prompt"
)

type testPrompt struct {
	prompt.Prompt

	input string
}

func (t *testPrompt) Input() (string, error) {
	return t.input, nil
}

func TestInteractiveProtoFiller(t *testing.T) {
	f := NewInteractiveFiller(nil, "")
	err := f.Fill("invalid type", fill.InteractiveFillerOpts{})
	if err != fill.ErrCodecMismatch {
		t.Errorf("must return fill.ErrCodecMismatch because the arg is invalid type, but got: %s", err)
	}

	tp := &testPrompt{
		input: "../../go.mod",
	}

	f = NewInteractiveFiller(tp, "")
	f.bytesFromFile = true

	var v interface{}
	v, err = f.inputPrimitiveField(descriptor.FieldDescriptorProto_TYPE_BYTES)
	if err != nil {
		t.Error(err)
	}

	if _, ok := v.([]byte); !ok {
		t.Errorf("value should be of type []byte")
	}

	fileContent, err := readFileFromRelativePath(tp.input)
	if err != nil {
		t.Error(err)
	}

	if len(v.([]byte)) != len(fileContent) {
		t.Error("contents should have the same length")
	}

	tp = &testPrompt{
		input: "\\x6f\\x67\\x69\\x73\\x6f",
	}

	f = NewInteractiveFiller(tp, "")

	v, err = f.inputPrimitiveField(descriptor.FieldDescriptorProto_TYPE_BYTES)
	if err != nil {
		t.Error(err)
	}

	if _, ok := v.([]byte); !ok {
		t.Errorf("value should be of type []byte")
	}

	if string(v.([]byte)) != "ogiso" {
		t.Error("unequal content")
	}
}
