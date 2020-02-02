package usecase

import (
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/pkg/errors"
)

func ListServices() ([]string, error) {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() ([]string, error) {
	return m.listServices(m.state.selectedPackage)
}

func (m *dependencyManager) listServices(pkgName string) ([]string, error) {
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
