package inputter

import (
	"encoding/json"
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/adapter/protobuf"
	"github.com/ktr0731/evans/entity"
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

func (i *JSONFile) Input(reqType entity.Message) (proto.Message, error) {
	req := protobuf.NewDynamicBuilder().NewMessage(reqType)
	err := i.decoder.Decode(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read input from JSON")
	}
	return req, nil
}
