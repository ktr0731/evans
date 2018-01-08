package usecase

import (
	"github.com/ktr0731/evans/entity"
	"github.com/ktr0731/evans/usecase/port"
)

func Package(params *port.PackageParams, outputPort port.OutputPort, logger port.Logger, env entity.Environment) (*port.PackageResponse, error) {
	if err := env.UsePackage(params.PkgName); err != nil {
		switch err {
		case entity.ErrUnknownPackage:
			logger.Printf("unknown package: %s")
		}
	}
	return outputPort.Package()
}
