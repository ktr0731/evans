package usecase

import "strings"

func GetDomainSourceName() string {
	return dm.GetDomainSourceName()
}
func (m *dependencyManager) GetDomainSourceName() string {
	var s []string
	if pkg := m.state.selectedPackage; pkg != "" {
		s = append(s, pkg)
	}
	if m.state.selectedService != "" {
		s = append(s, m.state.selectedService)
	}
	return strings.Join(s, ".")
}
