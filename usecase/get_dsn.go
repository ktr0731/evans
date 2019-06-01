package usecase

func GetDomainSourceName() string {
	return dm.GetDomainSourceName()
}
func (m *dependencyManager) GetDomainSourceName() (dsn string) {
	if m.state.selectedPackage != "" {
		dsn = m.state.selectedPackage
	}
	if m.state.selectedService != "" {
		dsn += "." + m.state.selectedService
	}
	return
}
