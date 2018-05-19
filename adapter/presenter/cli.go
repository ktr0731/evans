package presenter

import (
	"bytes"
	"io"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/usecase"
	"github.com/ktr0731/evans/usecase/port"
)

type marshaler interface {
	Marshal(io.Writer, proto.Message) error
}

type CLIPresenter struct {
	marshaler marshaler
}

func (p *CLIPresenter) Package() (io.Reader, error) {
	return nil, nil
}

func (p *CLIPresenter) Service() (io.Reader, error) {
	return nil, nil
}

func (p *CLIPresenter) Describe(showable port.Showable) (io.Reader, error) {
	return strings.NewReader(showable.Show()), nil
}

func (p *CLIPresenter) Show(showable port.Showable) (io.Reader, error) {
	return strings.NewReader(showable.Show()), nil
}

func (p *CLIPresenter) Header() (io.Reader, error) {
	return nil, nil
}

func (p *CLIPresenter) Call(res proto.Message) (io.Reader, error) {
	buf := new(bytes.Buffer)
	if err := p.marshaler.Marshal(buf, res); err != nil {
		return nil, err
	}
	if _, err := buf.WriteRune('\n'); err != nil {
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
	if ssr, ok := pb.(*usecase.ServerStreamingResult); ok {
		for i, res := range ssr.Res {
			if err := m.Marshal(out, res); err != nil {
				return err
			}
			if i == len(ssr.Res)-1 {
				io.WriteString(out, "\n")
			} else {
				io.WriteString(out, "\n\n")
			}
		}
		return nil
	}
	return m.Marshaler.Marshal(out, pb)
}

func (m *jsonMarshaler) hasIndent() bool {
	return m.Marshaler.Indent != ""
}
