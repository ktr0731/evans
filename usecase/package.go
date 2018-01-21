package usecase

import (
	"io"

	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Package(params *port.PackageParams, outputPort port.OutputPort, env entity.Environment) (io.Reader, error) {
	if err := env.UsePackage(params.PkgName); err != nil {
		return nil, err
	}
	return outputPort.Package()
}
