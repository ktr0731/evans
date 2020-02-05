package usecase

import "github.com/ktr0731/evans/idl"

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
	for _, pkg := range ListPackages() {
		if pkg == pkgName {
			m.state.selectedPackage = pkgName
			return nil
		}
	}
	return idl.ErrUnknownPackageName
}
