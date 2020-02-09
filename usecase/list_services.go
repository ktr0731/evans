package usecase

import (
	"github.com/ktr0731/evans/grpc/grpcreflection"
)

// ListServices returns the loaded fully-qualified service names.
func ListServices() []string {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() []string {
	return m.listServices()
}

func (m *dependencyManager) listServices() []string {
	result := m.spec.ServiceNames()
	for i := range result {
		if result[i] == grpcreflection.ServiceName {
			result = append(result[:i], result[i+1:]...)
			return result
		}
	}
	return result
}
