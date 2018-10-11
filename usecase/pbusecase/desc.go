package pbusecase

import (
	"io"

	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
)

func Describe(params *port.DescribeParams, outputPort port.OutputPort, env env.Environment) (io.Reader, error) {
	msg, err := env.Message(params.MsgName)
	if err != nil {
		return nil, err
	}
	return outputPort.Describe(&message{msg})
}
