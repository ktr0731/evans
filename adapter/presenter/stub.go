package presenter

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/ktr0731/evans/entity"
)

type StubPresenter struct{}

func (p *StubPresenter) Package() (io.Reader, error) {
	return nil, nil
}

func (p *StubPresenter) Service() (io.Reader, error) {
	return nil, nil
}

func (p *StubPresenter) Describe(msg *entity.Message) (io.Reader, error) {
	return nil, nil
}

func (p *StubPresenter) Show() (io.Reader, error) {
	return nil, nil
}

func (p *StubPresenter) Header() (io.Reader, error) {
	return nil, nil
}

func (p *StubPresenter) Call(res proto.Message) (io.Reader, error) {
	return p.marshal(res)
}

func (p *StubPresenter) marshal(pb proto.Message) (io.Reader, error) {
	if pb == nil {
		return nil, nil
	}

	if msg, ok := pb.(*dynamic.Message); ok {
		b, err := msg.MarshalJSON()
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	}

	buf := new(bytes.Buffer)
	marshaler := &jsonpb.Marshaler{}
	err := marshaler.Marshal(buf, pb)
	return buf, err
}

func NewStubPresenter() *StubPresenter {
	return &StubPresenter{}
}
