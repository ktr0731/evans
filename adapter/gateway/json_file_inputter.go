package gateway

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
	"github.com/pkg/errors"
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
	err := i.decoder.Decode(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read input from JSON")
	}
	return req, nil
}
