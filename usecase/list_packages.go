package usecase

import (
	"sort"

	"github.com/ktr0731/evans/idl/proto"
)

// ListPackages lists all package names.
func ListPackages() []string {
	return dm.ListPackages()
}
func (m *dependencyManager) ListPackages() []string {
	svcNames := m.spec.ServiceNames()
	encountered := make(map[string]interface{})
	toPackageName := func(svcName string) string {
		pkg, _ := proto.ParseFullyQualifiedServiceName(svcName)
		return pkg
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
