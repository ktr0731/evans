package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity/env"
	"github.com/ktr0731/evans/usecase/port"
)

func Package(params *port.PackageParams, outputPort port.OutputPort, env env.Environment) (io.Reader, error) {
	if err := env.UsePackage(params.PkgName); err != nil {
		return nil, err
	}
	return outputPort.Package()
}
