package usecase

// ListServices returns the loaded fully-qualified service names.
func ListServices() []string {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() []string {
	return m.listServices()
}

func (m *dependencyManager) listServices() []string {
	return m.descSource.ListServices()
}
