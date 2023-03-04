package usecase

// ListServices returns the loaded fully-qualified service names.
func ListServices() ([]string, error) {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() ([]string, error) {
	return m.listServices()
}

func (m *dependencyManager) listServices() ([]string, error) {
	return m.descSource.ListServices()
}
