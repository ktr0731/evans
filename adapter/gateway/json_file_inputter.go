package gateway

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
)

type JSONFileInputter struct {
	decoder *json.Decoder
}

func NewJSONFileInputter(in io.Reader) *JSONFileInputter {
	return &JSONFileInputter{
		decoder: json.NewDecoder(in),
	}
}

func (i *JSONFileInputter) Input(reqType entity.Message) (proto.Message, error) {
	req := protobuf.NewDynamicMessage(reqType)
	return req, i.decoder.Decode(req)
}
