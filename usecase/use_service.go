package usecase

import (
	"github.com/ktr0731/evans/proto"
	"github.com/pkg/errors"
)

// UseService modifies svcName as the currently selected service.
// UseService may return these errors:
//
//   - ErrPackageUnselected: REPL never call UsePackage.
//   - ErrUnknownServiceName: svcName is not in loaded services.
//
func UseService(svcName string) error {
	return dm.UseService(svcName)
}
func (m *dependencyManager) UseService(svcName string) error {
	if svcName == "" {
		return errors.Errorf("invalid service name '%s'", svcName)
	}
	var hasPackage bool
	for _, fqsn := range m.descSource.ListServices() {
		pkg, svc := proto.ParseFullyQualifiedServiceName(fqsn)
		if m.state.selectedPackage == pkg {
			hasPackage = true
			if svcName == svc {
				m.state.selectedService = svcName
				return nil
			}
		}
	}
	if hasPackage {
		return ErrUnknownServiceName
	}
	// In the case of empty package.
	return ErrPackageUnselected
}
