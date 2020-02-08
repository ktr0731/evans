package usecase

import (
	"strings"

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
	if svcName == "" {
		return errors.Errorf("invalid service name '%s'", svcName)
	}
	var hasPackage bool
	for _, fqsn := range m.spec.ServiceNames() {
		i := strings.LastIndex(fqsn, ".")
		var pkg, svc string
		if i == -1 {
			svc = fqsn
		} else {
			pkg, svc = fqsn[:i], fqsn[i+1:]
		}
		if m.state.selectedPackage == pkg {
			hasPackage = true
			if svcName == svc {
				m.state.selectedService = svcName
				return nil
			}
		}
	}
	if hasPackage {
		return idl.ErrUnknownServiceName
	}
	// In the case of empty package.
	return idl.ErrPackageUnselected
}
