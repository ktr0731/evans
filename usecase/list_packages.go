package usecase

import (
	"sort"
	"strings"
)

// ListPackages lists all package names.
func ListPackages() []string {
	return dm.ListPackages()
}
func (m *dependencyManager) ListPackages() []string {
	svcNames := m.spec.ServiceNames()
	encountered := make(map[string]interface{})
	toPackageName := func(svcName string) string {
		i := strings.LastIndex(svcName, ".")
		if i == -1 {
			return ""
		}
		return svcName[:i]
	}
	for _, svc := range svcNames {
		encountered[toPackageName(svc)] = nil
	}
	pkgs := make([]string, 0, len(svcNames))
	for pkg := range encountered {
		pkgs = append(pkgs, pkg)
	}

	sort.Slice(pkgs, func(i, j int) bool {
		return pkgs[i] < pkgs[j]
	})
	return pkgs
}
