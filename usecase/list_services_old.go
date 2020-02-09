package usecase

import (
	"github.com/ktr0731/evans/idl/proto"
)

// ListServicesOld returns the services belong to the selected package.
// The returned service names are NOT fully-qualified.
func ListServicesOld() []string {
	return dm.ListServicesOld()
}
func (m *dependencyManager) ListServicesOld() []string {
	return m.listServicesOld(m.state.selectedPackage)
}

func (m *dependencyManager) listServicesOld(pkgName string) []string {
	var svcs []string
	svcNames := m.spec.ServiceNames()
	for i := range svcNames {
		pkg, svc := proto.ParseFullyQualifiedServiceName(svcNames[i])
		if pkg == pkgName {
			svcs = append(svcs, svc)
		}
	}
	return svcs
}
