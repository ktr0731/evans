package usecase

import "strings"

func ListServices() []string {
	return dm.ListServices()
}
func (m *dependencyManager) ListServices() []string {
	return m.listServices(m.state.selectedPackage)
}

func (m *dependencyManager) listServices(pkgName string) []string {
	svcs := make([]string, 0, len(m.spec.ServiceNames()))
	for _, fqsn := range m.spec.ServiceNames() {
		i := strings.LastIndex(fqsn, ".")
		var pkg, svc string
		if i == -1 {
			svc = fqsn
		} else {
			pkg, svc = fqsn[:i], fqsn[i+1:]
		}
		if pkg == pkgName {
			svcs = append(svcs, svc)
		}
	}
	return svcs
}
