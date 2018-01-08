package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Describe(params *port.DescribeParams, outputPort port.OutputPort, env entity.Environment) (io.Reader, error) {
	msg, err := env.Message(params.MsgName)
	if err != nil {
		return nil, err
	}
	return outputPort.Describe(msg)
}
