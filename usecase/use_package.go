package usecase

// UsePackage modifies pkgName as the currently selected package.
// UsePackage may return these errors:
//
//   - ErrUnknownPackageName: pkgName is not in loaded packages.
//
func UsePackage(pkgName string) error {
	return dm.UsePackage(pkgName)
}
func (m *dependencyManager) UsePackage(pkgName string) error {
	pkgs, err := ListPackages()
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		if pkg == pkgName {
			m.state.selectedPackage = pkgName
			m.state.selectedService = ""
			return nil
		}
	}
	return ErrUnknownPackageName
}
