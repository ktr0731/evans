package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Package(params *port.PackageParams, outputPort port.OutputPort, env entity.Environment) (*port.PackageResponse, error) {
	if err := env.UsePackage(params.PkgName); err != nil {
		return nil, err
	}
	return outputPort.Package()
}
