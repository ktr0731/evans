package usecase

import (
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/pkg/errors"
)

func ListServicesOld() ([]string, error) {
	return dm.ListServicesOld()
}
func (m *dependencyManager) ListServicesOld() ([]string, error) {
	return m.listServicesOld(m.state.selectedPackage)
}

func (m *dependencyManager) listServicesOld(pkgName string) ([]string, error) {
	svcNames, err := m.spec.ServiceNames(pkgName)
	if err != nil {
		return nil, errors.Wrap(err, "invalid package name")
	}
	for i := range svcNames {
		if svcNames[i] == grpcreflection.ServiceName {
			return append(svcNames[:i], svcNames[i+1:]...), nil
		}
	}
	return svcNames, nil
}
