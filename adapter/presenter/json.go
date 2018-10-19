package presenter

import (
	"bytes"
	"io"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/usecase/port"
)

type JSONPresenter struct {
	indent string
}

var emptyResult = strings.NewReader("")

func (p *JSONPresenter) Package() (io.Reader, error) {
	return emptyResult, nil
}

func (p *JSONPresenter) Service() (io.Reader, error) {
	return emptyResult, nil
}

func (p *JSONPresenter) Describe(showable port.Showable) (io.Reader, error) {
	return strings.NewReader(showable.Show()), nil
}

func (p *JSONPresenter) Show(showable port.Showable) (io.Reader, error) {
	return strings.NewReader(showable.Show()), nil
}

func (p *JSONPresenter) Header() (io.Reader, error) {
	return emptyResult, nil
}

func (p *JSONPresenter) Call(res proto.Message) (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := marshalIndent(buf, res, p.indent); err != nil {
		return nil, err
	}
	if _, err := buf.WriteRune('\n'); err != nil {
		return nil, err
	}
	return buf, nil
}

func NewJSON() *JSONPresenter {
	return &JSONPresenter{}
}

func NewJSONWithIndent() *JSONPresenter {
	return &JSONPresenter{indent: "  "}
}

func marshalIndent(out io.Writer, pb proto.Message, indent string) error {
	if dmsg, ok := pb.(*dynamic.Message); ok {
		b, err := dmsg.MarshalJSONPB(&jsonpb.Marshaler{
			Indent:       indent,
			EmitDefaults: true,
		})
		if err != nil {
			return err
		}
		_, err = out.Write(b)
		return err
	} else {
		m := &jsonpb.Marshaler{}
		return m.Marshal(out, pb)
	}
}
