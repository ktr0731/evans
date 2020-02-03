package usecase

import (
	"github.com/ktr0731/evans/grpc/grpcreflection"
	"github.com/pkg/errors"
)

func ListServices() ([]string, error) {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() ([]string, error) {
	return m.listServices()
}

func (m *dependencyManager) listServices() ([]string, error) {
	var result []string
	for _, pkgName := range ListPackages() {
		svcNames, err := m.spec.ServiceNames(pkgName)
		if err != nil {
			return nil, errors.Wrap(err, "invalid package name")
		}
		result = append(result, svcNames...)
	}
	for i := range result {
		if result[i] == grpcreflection.ServiceName {
			result = append(result[:i], result[i+1:]...)
			return result, nil
		}
	}
	return result, nil
}
