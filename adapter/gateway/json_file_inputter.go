package gateway

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
)

type JSONFileInputter struct {
	decoder *json.Decoder
}

func NewJSONFileInputter(in io.Reader) *JSONFileInputter {
	return &JSONFileInputter{
		decoder: json.NewDecoder(in),
	}
}

func (i *JSONFileInputter) Input(reqType *desc.MessageDescriptor) (proto.Message, error) {
	req := dynamic.NewMessage(reqType)
	return req, i.decoder.Decode(req)
}
