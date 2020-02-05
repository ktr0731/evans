package usecase

import (
	"github.com/ktr0731/evans/idl"
	"github.com/pkg/errors"
)

// UseService modifies svcName as the currently selected service.
// UseService may return these errors:
//
//   - ErrPackageUnselected: REPL never call UsePackage.
//   - ErrUnknownServiceName: svcName is not in loaded services.
//   - Other errors.
//
func UseService(svcName string) error {
	return dm.UseService(svcName)
}
func (m *dependencyManager) UseService(svcName string) error {
	for _, pkg := range ListPackages() {
		if pkg == m.state.selectedPackage {
			_, err := m.listRPCs(m.state.selectedPackage, svcName)
			if err == idl.ErrServiceUnselected {
				return errors.Errorf("invalid service name '%s'", svcName)
			}
			if err != nil {
				return errors.Wrapf(err, "cannot use service '%s'", svcName)
			}
			m.state.selectedService = svcName
			return nil
		}
	}
	return idl.ErrPackageUnselected
}
