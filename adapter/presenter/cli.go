package presenter

import (
	"bytes"
	"io"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

type marshaler interface {
	Marshal(io.Writer, proto.Message) error
}

type CLIPresenter struct {
	marshaler marshaler
}

func (p *CLIPresenter) Package() (*port.PackageResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Service() (*port.ServiceResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Describe(msg *entity.Message) (io.Reader, error) {
	return strings.NewReader(msg.String()), nil
}

func (p *CLIPresenter) Show() (*port.ShowResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Header() (*port.HeaderResponse, error) {
	return nil, nil
}

func (p *CLIPresenter) Call(res proto.Message) (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := p.marshaler.Marshal(buf, res); err != nil {
		return nil, err
	}
	return buf, nil
}

func NewJSONCLIPresenter() *CLIPresenter {
	return &CLIPresenter{
		marshaler: &jsonMarshaler{jsonpb.Marshaler{}},
	}
}

func NewJSONCLIPresenterWithIndent() *CLIPresenter {
	presenter := NewJSONCLIPresenter()
	presenter.marshaler.(*jsonMarshaler).Indent = "  "
	return presenter
}

type jsonMarshaler struct {
	jsonpb.Marshaler
}

func (m *jsonMarshaler) Marshal(out io.Writer, pb proto.Message) error {
	if dmsg, ok := pb.(*dynamic.Message); ok {
		var b []byte
		var err error
		if m.hasIndent() {
			b, err = dmsg.MarshalJSONIndent()
		} else {
			b, err = dmsg.MarshalJSON()
		}
		if err != nil {
			return err
		}
		_, err = out.Write(b)
		return err
	}
	return m.Marshaler.Marshal(out, pb)
}

func (m *jsonMarshaler) hasIndent() bool {
	return m.Marshaler.Indent != ""
}
