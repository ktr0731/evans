package usecasetest

import (
	"io"

	"github.com/golang/protobuf/proto"
	"github.com/ktr0731/evans/tests/mock/usecase/mockport"
	"github.com/ktr0731/evans/usecase/port"
)

func NewPresenter() port.OutputPort {
	return &mockport.OutputPortMock{
		DescribeFunc: func(showable port.Showable) (io.Reader, error) {
			return nil, nil
		},
		HeaderFunc: func() (io.Reader, error) {
			return nil, nil
		},
		PackageFunc: func() (io.Reader, error) {
			return nil, nil
		},
		ServiceFunc: func() (io.Reader, error) {
			return nil, nil
		},
		ShowFunc: func(showable port.Showable) (io.Reader, error) {
			return nil, nil
		},
		CallFunc: func(res proto.Message) (io.Reader, error) {
			return nil, nil
		},
	}
}
