package inputter

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

type JSONFile struct {
	decoder *json.Decoder
}

func NewJSONFile(in io.Reader) *JSONFile {
	return &JSONFile{
		decoder: json.NewDecoder(in),
	}
}

func (i *JSONFile) Input(req *desc.MessageDescriptor) (proto.Message, error) {
	dmsg := dynamic.NewMessage(req)
	err := i.decoder.Decode(dmsg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read input from JSON")
	}
	return dmsg, nil
}
