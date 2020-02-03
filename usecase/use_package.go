package usecase

import (
	"github.com/ktr0731/evans/idl"
	"github.com/pkg/errors"
)

// UsePackage modifies pkgName as the currently selected package.
// UsePackage may return these errors:
//
//   - idl.ErrUnknownPackageName: pkgName is not in loaded packages.
//   - Other errors.
//
func UsePackage(pkgName string) error {
	return dm.UsePackage(pkgName)
}
func (m *dependencyManager) UsePackage(pkgName string) error {
	_, err := m.listServicesOld(pkgName)
	if err == idl.ErrPackageUnselected {
		return errors.Errorf("invalid package name '%s'", pkgName)
	}
	if err != nil {
		return errors.Wrapf(err, "cannot use package '%s'", pkgName)
	}
	m.state.selectedPackage = pkgName
	return nil
}
