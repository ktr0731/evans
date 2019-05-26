package usecase

// ListPackages lists all package names.
func ListPackages() []string {
	return dm.ListPackages()
}
func (m *dependencyManager) ListPackages() []string {
	return m.spec.PackageNames()
}
